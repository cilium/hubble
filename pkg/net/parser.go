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

package net

import (
	"fmt"
	"net"
	"strings"
)

// ParseIP converts an IP address from a string into a either a IPv4 or a IPv6
// net.IP. If the given 'ip' address is not valid, an error is returned.
func ParseIP(ip string) (ipv4s, ipv6s net.IP, err error) {
	addr := net.ParseIP(ip)
	if addr == nil {
		return nil, nil, fmt.Errorf("invalid ip address: %s", ip)
	}

	// apparently, `ip.To4() != nil` is not reliable enough to check if a net.IP
	// is really an IPv4 address, so we inspect the string. See asaskevich/govalidator#100.
	if strings.Contains(ip, ":") {
		return nil, addr, nil
	}
	return addr.To4(), nil, nil
}

// ParseIPs returns a slice of IPv4 and IPv6s. If any of the given IP strings
// are not a valid IP address, an error is returned.
func ParseIPs(ips []string) (ipv4s, ipv6s []net.IP, err error) {
	for _, ip := range ips {
		v4, v6, err := ParseIP(ip)
		switch {
		case err != nil:
			return nil, nil, err
		case v4 != nil:
			ipv4s = append(ipv4s, v4)
		case v6 != nil:
			ipv6s = append(ipv6s, v6)
		}
	}

	return ipv4s, ipv6s, nil
}
