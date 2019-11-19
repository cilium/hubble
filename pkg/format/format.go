// Copyright 2019 Authors of Hubble
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

package format

import (
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket/layers"
)

var (
	// EnablePortTranslation enables translation of port numbers to port
	// names, i.e. `80` becomes `80(http)`.
	EnablePortTranslation = true

	// EnableIPTranslation enables translation of IPs to pod names, FQDNs,
	// service names ...
	EnableIPTranslation = true

	// additional named ports that are related to hubble and cilium, which do
	// not appear in the "well known" list of Golang ports... yet.
	namedPorts = map[int]string{
		4240: "cilium-health",
	}
)

// MaybeTime returns a Millisecond precision timestamp, or "N/A" if nil.
func MaybeTime(t *time.Time) string {
	if t != nil {
		// TODO: support more date formats through options to `hubble observe`
		return t.Format(time.StampMilli)
	}
	return "N/A"
}

// UDPPort ...
func UDPPort(p layers.UDPPort) string {
	i := int(p)
	if !EnablePortTranslation {
		return strconv.Itoa(i)
	}
	if name, ok := namedPorts[i]; ok {
		return fmt.Sprintf("%v(%v)", i, name)
	}
	return p.String()
}

// TCPPort ...
func TCPPort(p layers.TCPPort) string {
	i := int(p)
	if !EnablePortTranslation {
		return strconv.Itoa(i)
	}
	if name, ok := namedPorts[i]; ok {
		return fmt.Sprintf("%v(%v)", i, name)
	}
	return p.String()
}

// Hostname returns a "host:ip" formatted pair for the given ip and port. If
// port is empty, only the host is returned. The host part is either the pod
// name (if set), or a comman-separated list of domain names (if set), or just
// the ip address if EnableIPTranslation is false and/or there are no pod and
// domain names.
func Hostname(ip, port string, ns, pod string, names []string) (host string) {
	host = ip
	if EnableIPTranslation {
		if pod != "" {
			// path.Join omits the slash if ns is empty
			host = path.Join(ns, pod)
		} else if len(names) != 0 {
			host = strings.Join(names, ",")
		}
	}

	if port != "" {
		return net.JoinHostPort(host, port)
	}

	return host
}
