module github.com/cilium/hubble

go 1.13

require (
	github.com/cilium/cilium v1.7.0-rc2
	github.com/fatih/color v1.7.1-0.20181010231311-3f9d52f7176a // indirect
	github.com/francoispqt/gojay v1.2.13
	github.com/go-openapi/strfmt v0.19.3
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/google/gopacket v1.1.17
	github.com/google/gops v0.3.6
	github.com/hashicorp/golang-lru v0.5.3
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/miekg/dns v1.1.22 // indirect
	github.com/mitchellh/protoc-gen-go-json v0.0.0-20200113165135-fd297ce346f1
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v1.2.0
	github.com/sasha-s/go-deadlock v0.2.1-0.20190130213442-5dc88f41ca59 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	google.golang.org/genproto v0.0.0-20190801165951-fa694d86fc64 // indirect
	google.golang.org/grpc v1.23.1
	k8s.io/apimachinery v0.17.1
)

replace (
	github.com/optiopay/kafka => github.com/cilium/kafka v0.0.0-20180809090225-01ce283b732b
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200111014337-d224476cd073
	k8s.io/apiextensions-apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200111014337-d224476cd073
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200111014337-d224476cd073
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200111014337-d224476cd073
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200111014337-d224476cd073
	k8s.io/client-go => github.com/cilium/client-go v0.0.0-20200116081620-3539db8110dd
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200111014337-d224476cd073
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200111014337-d224476cd073
	k8s.io/code-generator => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20200111014337-d224476cd073
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200111014337-d224476cd073
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200111014337-d224476cd073
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200111014337-d224476cd073
	k8s.io/gengo => k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/heapster => k8s.io/heapster v1.2.0-beta.1
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/kube-aggregator => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20200111014337-d224476cd073
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200111014337-d224476cd073
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200111014337-d224476cd073
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200111014337-d224476cd073
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20200111014337-d224476cd073
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200111014337-d224476cd073
	k8s.io/kubernetes => k8s.io/kubernetes v1.17.1
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200111014337-d224476cd073
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200111014337-d224476cd073
	k8s.io/repo-infra => k8s.io/repo-infra v0.0.1-alpha.1
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200111014337-d224476cd073
	k8s.io/system-validators => k8s.io/system-validators v1.0.4
	k8s.io/utils => k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
)
