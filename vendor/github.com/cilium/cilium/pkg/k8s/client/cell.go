// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cilium/hive"
	"github.com/cilium/hive/cell"
	"github.com/cilium/hive/script"
	"github.com/sirupsen/logrus"
	apiext_clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiext_fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	versionapi "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8sTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/connrotation"
	mcsapi_clientset "sigs.k8s.io/mcs-api/pkg/client/clientset/versioned"
	mcsapi_fake "sigs.k8s.io/mcs-api/pkg/client/clientset/versioned/fake"

	"github.com/cilium/cilium/pkg/controller"
	cilium_clientset "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	cilium_fake "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/fake"
	k8smetrics "github.com/cilium/cilium/pkg/k8s/metrics"
	slim_apiextclientsetscheme "github.com/cilium/cilium/pkg/k8s/slim/k8s/apiextensions-client/clientset/versioned/scheme"
	slim_apiext_clientset "github.com/cilium/cilium/pkg/k8s/slim/k8s/apiextensions-clientset"
	slim_metav1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	slim_metav1beta1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1beta1"
	slim_clientset "github.com/cilium/cilium/pkg/k8s/slim/k8s/client/clientset/versioned"
	slim_fake "github.com/cilium/cilium/pkg/k8s/slim/k8s/client/clientset/versioned/fake"
	"github.com/cilium/cilium/pkg/k8s/testutils"
	k8sversion "github.com/cilium/cilium/pkg/k8s/version"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/version"
)

// client.Cell provides Clientset, a composition of clientsets to Kubernetes resources
// used by Cilium.
var Cell = cell.Module(
	"k8s-client",
	"Kubernetes Client",

	cell.Config(defaultSharedConfig),
	cell.Config(defaultClientParams),
	cell.Provide(NewClientConfig),
	cell.Provide(newClientset),
)

// client.ClientBuilderCell provides a function to create a new composite Clientset,
// allowing a controller to use its own Clientset with a different user agent.
var ClientBuilderCell = cell.Module(
	"k8s-client-builder",
	"Kubernetes Client Builder",

	cell.Config(defaultSharedConfig),
	cell.Provide(NewClientConfig),
	cell.Provide(NewClientBuilder),
)

var k8sHeartbeatControllerGroup = controller.NewGroup("k8s-heartbeat")

// Type aliases for the clientsets to avoid name collision on 'Clientset' when composing them.
type (
	MCSAPIClientset     = mcsapi_clientset.Clientset
	KubernetesClientset = kubernetes.Clientset
	SlimClientset       = slim_clientset.Clientset
	APIExtClientset     = slim_apiext_clientset.Clientset
	CiliumClientset     = cilium_clientset.Clientset
)

// Clientset is a composition of the different client sets used by Cilium.
type Clientset interface {
	mcsapi_clientset.Interface
	kubernetes.Interface
	apiext_clientset.Interface
	cilium_clientset.Interface
	Getters

	// Slim returns the slim client, which contains some of the same APIs as the
	// normal kubernetes client, but with slimmed down messages to reduce memory
	// usage. Prefer the slim version when caching messages.
	Slim() slim_clientset.Interface

	// IsEnabled returns true if Kubernetes support is enabled and the
	// clientset can be used.
	IsEnabled() bool

	// Disable disables the client. Panics if called after the clientset has been
	// started.
	Disable()

	// Config returns the configuration used to create this client.
	Config() Config

	// RestConfig returns the deep copy of rest configuration.
	RestConfig() *rest.Config
}

// compositeClientset implements the Clientset using real clients.
type compositeClientset struct {
	started  bool
	disabled bool

	*MCSAPIClientset
	*KubernetesClientset
	*APIExtClientset
	*CiliumClientset
	clientsetGetters

	controller    *controller.Manager
	slim          *SlimClientset
	config        Config
	log           logrus.FieldLogger
	closeAllConns func()
	restConfig    *rest.Config
}

func newClientset(lc cell.Lifecycle, log logrus.FieldLogger, cfg Config) (Clientset, error) {
	return newClientsetForUserAgent(lc, log, cfg, "")
}

func newClientsetForUserAgent(lc cell.Lifecycle, log logrus.FieldLogger, cfg Config, name string) (Clientset, error) {
	if !cfg.isEnabled() {
		return &compositeClientset{disabled: true}, nil
	}

	if cfg.K8sAPIServer != "" &&
		!strings.HasPrefix(cfg.K8sAPIServer, "http") {
		cfg.K8sAPIServer = "http://" + cfg.K8sAPIServer // default to HTTP
	}

	client := compositeClientset{
		log:        log,
		controller: controller.NewManager(),
		config:     cfg,
	}

	cmdName := "cilium"
	if len(os.Args[0]) != 0 {
		cmdName = filepath.Base(os.Args[0])
	}
	userAgent := fmt.Sprintf("%s/%s", cmdName, version.Version)

	if name != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, name)
	}

	restConfig, err := createConfig(cfg.K8sAPIServer, cfg.K8sKubeConfigPath, cfg.K8sClientQPS, cfg.K8sClientBurst, userAgent)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s client rest configuration: %w", err)
	}
	client.restConfig = restConfig
	defaultCloseAllConns := setDialer(cfg, restConfig)

	httpClient, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s REST client: %w", err)
	}

	// We are implementing the same logic as Kubelet, see
	// https://github.com/kubernetes/kubernetes/blob/v1.24.0-beta.0/cmd/kubelet/app/server.go#L852.
	if s := os.Getenv("DISABLE_HTTP2"); len(s) > 0 {
		client.closeAllConns = defaultCloseAllConns
	} else {
		client.closeAllConns = func() {
			utilnet.CloseIdleConnectionsFor(restConfig.Transport)
		}
	}

	// Slim and K8s clients use protobuf marshalling.
	restConfig.ContentConfig.ContentType = `application/vnd.kubernetes.protobuf`

	client.slim, err = slim_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create slim k8s client: %w", err)
	}

	client.APIExtClientset, err = slim_apiext_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create apiext k8s client: %w", err)
	}

	client.MCSAPIClientset, err = mcsapi_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create mcsapi k8s client: %w", err)
	}

	client.KubernetesClientset, err = kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create k8s client: %w", err)
	}

	client.clientsetGetters = clientsetGetters{&client}

	// The cilium client uses JSON marshalling.
	restConfig.ContentConfig.ContentType = `application/json`
	client.CiliumClientset, err = cilium_clientset.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, fmt.Errorf("unable to create cilium k8s client: %w", err)
	}

	lc.Append(cell.Hook{
		OnStart: client.onStart,
		OnStop:  client.onStop,
	})

	return &client, nil
}

func (c *compositeClientset) Slim() slim_clientset.Interface {
	return c.slim
}

func (c *compositeClientset) Discovery() discovery.DiscoveryInterface {
	return c.KubernetesClientset.Discovery()
}

func (c *compositeClientset) IsEnabled() bool {
	return c != nil && c.config.isEnabled() && !c.disabled
}

func (c *compositeClientset) Disable() {
	if c.started {
		panic("Clientset.Disable() called after it had been started")
	}
	c.disabled = true
}

func (c *compositeClientset) Config() Config {
	return c.config
}

func (c *compositeClientset) RestConfig() *rest.Config {
	return rest.CopyConfig(c.restConfig)
}

func (c *compositeClientset) onStart(startCtx cell.HookContext) error {
	if !c.IsEnabled() {
		return nil
	}

	if err := c.waitForConn(startCtx); err != nil {
		return err
	}
	c.startHeartbeat()

	// Update the global K8s clients, K8s version and the capabilities.
	if err := k8sversion.Update(c, c.config.EnableK8sAPIDiscovery); err != nil {
		return err
	}

	if !k8sversion.Capabilities().MinimalVersionMet {
		return fmt.Errorf("k8s version (%v) is not meeting the minimal requirement (%v)",
			k8sversion.Version(), k8sversion.MinimalVersionConstraint)
	}

	c.started = true

	return nil
}

func (c *compositeClientset) onStop(stopCtx cell.HookContext) error {
	if c.IsEnabled() {
		c.controller.RemoveAllAndWait()
		c.closeAllConns()
	}
	c.started = false
	return nil
}

func (c *compositeClientset) startHeartbeat() {
	restClient := c.KubernetesClientset.RESTClient()

	timeout := c.config.K8sHeartbeatTimeout
	if timeout == 0 {
		return
	}

	heartBeat := func(ctx context.Context) error {
		// Kubernetes does a get node of the node that kubelet is running [0]. This seems excessive in
		// our case because the amount of data transferred is bigger than doing a Get of /healthz.
		// For this reason we have picked to perform a get on `/healthz` instead a get of a node.
		//
		// [0] https://github.com/kubernetes/kubernetes/blob/v1.17.3/pkg/kubelet/kubelet_node_status.go#L423
		res := restClient.Get().Resource("healthz").Do(ctx)
		return res.Error()
	}

	c.controller.UpdateController("k8s-heartbeat",
		controller.ControllerParams{
			Group: k8sHeartbeatControllerGroup,
			DoFunc: func(context.Context) error {
				runHeartbeat(
					c.log,
					heartBeat,
					timeout,
					c.closeAllConns,
				)
				return nil
			},
			RunInterval: timeout,
		})
}

// createConfig creates a rest.Config for connecting to k8s api-server.
//
// The precedence of the configuration selection is the following:
// 1. kubeCfgPath
// 2. apiServerURL (https if specified)
// 3. rest.InClusterConfig().
func createConfig(apiServerURL, kubeCfgPath string, qps float32, burst int, userAgent string) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	switch {
	// If the apiServerURL and the kubeCfgPath are empty then we can try getting
	// the rest.Config from the InClusterConfig
	case apiServerURL == "" && kubeCfgPath == "":
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	case kubeCfgPath != "":
		if config, err = clientcmd.BuildConfigFromFlags("", kubeCfgPath); err != nil {
			return nil, err
		}
	case strings.HasPrefix(apiServerURL, "https://"):
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
		config.Host = apiServerURL
	default:
		//exhaustruct:ignore
		config = &rest.Config{Host: apiServerURL, UserAgent: userAgent}
	}

	setConfig(config, userAgent, qps, burst)
	return config, nil
}

func setConfig(config *rest.Config, userAgent string, qps float32, burst int) {
	if userAgent != "" {
		config.UserAgent = userAgent
	}
	if qps != 0.0 {
		config.QPS = qps
	}
	if burst != 0 {
		config.Burst = burst
	}
}

func (c *compositeClientset) waitForConn(ctx context.Context) error {
	stop := make(chan struct{})
	timeout := time.NewTimer(time.Minute)
	defer timeout.Stop()
	var err error
	wait.Until(func() {
		c.log.WithField("host", c.restConfig.Host).Info("Establishing connection to apiserver")
		err = isConnReady(c)
		if err == nil {
			close(stop)
			return
		}

		select {
		case <-ctx.Done():
		case <-timeout.C:
		default:
			return
		}

		c.log.WithError(err).WithField(logfields.IPAddr, c.restConfig.Host).Error("Unable to contact k8s api-server")
		close(stop)
	}, 5*time.Second, stop)
	if err == nil {
		c.log.Info("Connected to apiserver")
	}
	return err
}

func setDialer(cfg Config, restConfig *rest.Config) func() {
	if cfg.K8sClientConnectionTimeout == 0 || cfg.K8sClientConnectionKeepAlive == 0 {
		return func() {}
	}
	ctx := (&net.Dialer{
		Timeout:   cfg.K8sClientConnectionTimeout,
		KeepAlive: cfg.K8sClientConnectionKeepAlive,
	}).DialContext
	dialer := connrotation.NewDialer(ctx)
	restConfig.Dial = dialer.DialContext
	return dialer.CloseAll
}

func runHeartbeat(log logrus.FieldLogger, heartBeat func(context.Context) error, timeout time.Duration, closeAllConns ...func()) {
	expireDate := time.Now().Add(-timeout)
	// Don't even perform a health check if we have received a successful
	// k8s event in the last 'timeout' duration
	if k8smetrics.LastSuccessInteraction.Time().After(expireDate) {
		return
	}

	done := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	go func() {
		// If we have reached up to this point to perform a heartbeat to
		// kube-apiserver then we should close the connections if we receive
		// any error at all except if we receive a http.StatusTooManyRequests
		// which means the server is overloaded and only for this reason we
		// will not close all connections.
		err := heartBeat(ctx)
		if err != nil {
			statusError := &k8sErrors.StatusError{}
			if !errors.As(err, &statusError) ||
				statusError.ErrStatus.Code != http.StatusTooManyRequests {
				done <- err
			}
		}
		close(done)
	}()

	select {
	case err := <-done:
		if err != nil {
			log.WithError(err).Warn("Network status error received, restarting client connections")
			for _, fn := range closeAllConns {
				fn()
			}
		}
	case <-ctx.Done():
		log.Warn("Heartbeat timed out, restarting client connections")
		for _, fn := range closeAllConns {
			fn()
		}
	}
}

// isConnReady returns the err for the kube-system namespace get
func isConnReady(c kubernetes.Interface) error {
	_, err := c.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	return err
}

var FakeClientCell = cell.Module(
	"k8s-fake-client",
	"Fake Kubernetes client",

	cell.Provide(
		NewFakeClientset,
		func(fc *FakeClientset) hive.ScriptCmdOut {
			return hive.NewScriptCmd("k8s", FakeClientCommand(fc))
		},
	),
)

type (
	MCSAPIFakeClientset     = mcsapi_fake.Clientset
	KubernetesFakeClientset = fake.Clientset
	SlimFakeClientset       = slim_fake.Clientset
	CiliumFakeClientset     = cilium_fake.Clientset
	APIExtFakeClientset     = apiext_fake.Clientset
)

type FakeClientset struct {
	disabled bool

	*MCSAPIFakeClientset
	*KubernetesFakeClientset
	*CiliumFakeClientset
	*APIExtFakeClientset
	clientsetGetters

	SlimFakeClientset *SlimFakeClientset

	trackers map[string]k8sTesting.ObjectTracker

	enabled bool
}

var _ Clientset = &FakeClientset{}

func (c *FakeClientset) Slim() slim_clientset.Interface {
	return c.SlimFakeClientset
}

func (c *FakeClientset) Discovery() discovery.DiscoveryInterface {
	return c.KubernetesFakeClientset.Discovery()
}

func (c *FakeClientset) IsEnabled() bool {
	return !c.disabled
}

func (c *FakeClientset) Disable() {
	c.disabled = true
}

func (c *FakeClientset) Config() Config {
	//exhaustruct:ignore
	return Config{}
}

func (c *FakeClientset) RestConfig() *rest.Config {
	//exhaustruct:ignore
	return &rest.Config{}
}

func NewFakeClientset() (*FakeClientset, Clientset) {
	version := testutils.DefaultVersion
	return NewFakeClientsetWithVersion(version)
}

func NewFakeClientsetWithVersion(version string) (*FakeClientset, Clientset) {
	if version == "" {
		version = testutils.DefaultVersion
	}
	resources, found := testutils.APIResources[version]
	if !found {
		panic("version " + version + " not found from testutils.APIResources")
	}

	client := FakeClientset{
		SlimFakeClientset:       slim_fake.NewSimpleClientset(),
		CiliumFakeClientset:     cilium_fake.NewSimpleClientset(),
		APIExtFakeClientset:     apiext_fake.NewSimpleClientset(),
		MCSAPIFakeClientset:     mcsapi_fake.NewSimpleClientset(),
		KubernetesFakeClientset: fake.NewSimpleClientset(),
		enabled:                 true,
	}
	client.KubernetesFakeClientset.Resources = resources
	client.SlimFakeClientset.Resources = resources
	client.CiliumFakeClientset.Resources = resources
	client.APIExtFakeClientset.Resources = resources
	client.trackers = map[string]k8sTesting.ObjectTracker{
		"slim":       client.SlimFakeClientset.Tracker(),
		"cilium":     client.CiliumFakeClientset.Tracker(),
		"mcs":        client.MCSAPIFakeClientset.Tracker(),
		"kubernetes": client.KubernetesFakeClientset.Tracker(),
		"apiexit":    client.APIExtFakeClientset.Tracker(),
	}

	fd := client.KubernetesFakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fd.FakedServerVersion = toVersionInfo(version)

	client.clientsetGetters = clientsetGetters{&client}
	return &client, &client
}

func toVersionInfo(rawVersion string) *versionapi.Info {
	parts := strings.Split(rawVersion, ".")
	return &versionapi.Info{Major: parts[0], Minor: parts[1]}
}

type ClientBuilderFunc func(name string) (Clientset, error)

// NewClientBuilder returns a function that creates a new Clientset with the given
// name appended to the user agent, or returns an error if the Clientset cannot be
// created.
func NewClientBuilder(lc cell.Lifecycle, log logrus.FieldLogger, cfg Config) ClientBuilderFunc {
	return func(name string) (Clientset, error) {
		c, err := newClientsetForUserAgent(lc, log, cfg, name)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

var FakeClientBuilderCell = cell.Provide(FakeClientBuilder)

func FakeClientBuilder() ClientBuilderFunc {
	fc, _ := NewFakeClientset()
	return func(_ string) (Clientset, error) {
		return fc, nil
	}
}

func FakeClientCommand(fc *FakeClientset) script.Cmd {
	return script.Command(
		script.CmdUsage{
			Summary: "interact with fake k8s client",
			Args:    "<command> args...",
		},
		func(s *script.State, args ...string) (script.WaitFunc, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("usage: k8s <command> files...\n<command> is one of add, update or delete.")
			}

			action := args[0]
			if len(args) < 2 {
				return nil, fmt.Errorf("usage: k8s %s files...", action)
			}

			for _, file := range args[1:] {
				b, err := os.ReadFile(s.Path(file))
				if err != nil {
					// Try relative to current directory, e.g. to allow reading "testdata/foo.yaml"
					b, err = os.ReadFile(file)
				}
				if err != nil {
					return nil, fmt.Errorf("failed to read %s: %w", file, err)
				}
				obj, gvk, err := testutils.DecodeObjectGVK(b)
				if err != nil {
					return nil, fmt.Errorf("decode: %w", err)
				}
				gvr, _ := meta.UnsafeGuessKindToResource(*gvk)
				objMeta, err := meta.Accessor(obj)
				if err != nil {
					return nil, fmt.Errorf("accessor: %w", err)
				}
				name := objMeta.GetName()
				ns := objMeta.GetNamespace()

				// Try to add the object to all the trackers. If one of them
				// accepts we're good. We'll add to all since multiple trackers
				// may accept (e.g. slim and kubernetes).

				// err will get set to nil if any of the tracker methods succeed.
				// start with a non-nil default error.
				err = fmt.Errorf("none of the trackers of FakeClientset accepted %T", obj)
				for trackerName, tracker := range fc.trackers {
					var trackerErr error
					switch action {
					case "add":
						trackerErr = tracker.Add(obj)
					case "update":
						trackerErr = tracker.Update(gvr, obj, ns)
					case "delete":
						trackerErr = tracker.Delete(gvr, ns, name)
					default:
						return nil, fmt.Errorf("unknown k8s action %q, expected 'add', 'update' or 'delete'", action)
					}
					if err != nil {
						if trackerErr == nil {
							// One of the trackers accepted the object, it's a success!
							err = nil
						} else {
							err = errors.Join(err, fmt.Errorf("%s: %w", trackerName, trackerErr))
						}
					}
				}
				if err != nil {
					return nil, err
				}
			}
			return nil, nil
		})
}

func init() {
	// Register the metav1.Table and metav1.PartialObjectMetadata for the
	// apiextclientset.
	utilruntime.Must(slim_metav1.AddMetaToScheme(slim_apiextclientsetscheme.Scheme))
	utilruntime.Must(slim_metav1beta1.AddMetaToScheme(slim_apiextclientsetscheme.Scheme))
}
