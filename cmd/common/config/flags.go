// Copyright 2020 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/pflag"
)

// Keys can be used to retrieve values from GlobalFlags and ServerFlags (e.g.
// when bound to a viper instance).
const (
	// GlobalFlags keys.
	KeyConfig = "config"
	KeyDebug  = "debug"

	// ServerFlags keys.
	KeyServer            = "server"
	KeyTLS               = "tls"
	KeyTLSAllowInsecure  = "tls-allow-insecure"
	KeyTLSCACertFiles    = "tls-ca-cert-files"
	KeyTLSClientCertFile = "tls-client-cert-file"
	KeyTLSClientKeyFile  = "tls-client-key-file"
	KeyTLSServerName     = "tls-server-name"
	KeyTimeout           = "timeout"
)

// GlobalFlags are flags that apply to any command.
var GlobalFlags = pflag.NewFlagSet("global", pflag.ContinueOnError)

// ServerFlags are flags that configure how to connect to a Hubble server.
var ServerFlags = pflag.NewFlagSet("server", pflag.ContinueOnError)

func init() {
	initGlobalFlags()
	initServerFlags()
}

func initGlobalFlags() {
	GlobalFlags.String(KeyConfig, defaults.ConfigFile, "Optional config file")
	GlobalFlags.BoolP(KeyDebug, "D", false, "Enable debug messages")
}

func initServerFlags() {
	ServerFlags.String(KeyServer, defaults.GetSocketPath(), "Address of a Hubble server")
	ServerFlags.Duration(KeyTimeout, defaults.DialTimeout, "Hubble server dialing timeout")
	ServerFlags.Bool(
		KeyTLS,
		false,
		"Specify that TLS must be used when establishing a connection to a Hubble server.\r\n"+
			"By default, TLS is only enabled if the server address starts with 'tls://'.",
	)
	ServerFlags.Bool(
		KeyTLSAllowInsecure,
		false,
		"Allows the client to skip verifying the server's certificate chain and host name.\r\n"+
			"This option is NOT recommended as, in this mode, TLS is susceptible to machine-in-the-middle attacks.\r\n"+
			"See also the 'tls-server-name' option which allows setting the server name.",
	)
	ServerFlags.StringSlice(
		KeyTLSCACertFiles,
		nil,
		"Paths to custom Certificate Authority (CA) certificate files."+
			"The files must contain PEM encoded data.",
	)
	ServerFlags.String(
		KeyTLSClientCertFile,
		"",
		"Path to the public key file for the client certificate to connect to a Hubble server (implies TLS).\r\n"+
			"The file must contain PEM encoded data.",
	)
	ServerFlags.String(
		KeyTLSClientKeyFile,
		"",
		"Path to the private key file for the client certificate to connect a Hubble server (implies TLS).\r\n"+
			"The file must contain PEM encoded data.",
	)
	ServerFlags.String(
		KeyTLSServerName,
		"",
		"Specify a server name to verify the hostname on the returned certificate (eg: 'instance.hubble-relay.cilium.io').",
	)
}
