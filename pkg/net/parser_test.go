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
	"net"
	"reflect"
	"testing"
)

func TestParseIP(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name      string
		args      args
		wantIpv4s net.IP
		wantIpv6s net.IP
		wantErr   bool
	}{
		{
			name:      "test 1",
			args:      args{ip: "1.1.1.1"},
			wantIpv4s: net.ParseIP("1.1.1.1").To4(),
			wantErr:   false,
		},
		{
			name:      "test 2",
			args:      args{ip: "fd00::1"},
			wantIpv6s: net.ParseIP("fd00::1"),
			wantErr:   false,
		},
		{
			name:    "test 3",
			args:    args{ip: "fdx0::1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIpv4s, gotIpv6s, err := ParseIP(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIpv4s, tt.wantIpv4s) {
				t.Errorf("ParseIP() gotIpv4s = %v, want %v", gotIpv4s, tt.wantIpv4s)
			}
			if !reflect.DeepEqual(gotIpv6s, tt.wantIpv6s) {
				t.Errorf("ParseIP() gotIpv6s = %v, want %v", gotIpv6s, tt.wantIpv6s)
			}
		})
	}
}

func TestParseIPs(t *testing.T) {
	type args struct {
		ips []string
	}
	tests := []struct {
		name      string
		args      args
		wantIpv4s []net.IP
		wantIpv6s []net.IP
		wantErr   bool
	}{
		{
			name: "",
			args: args{
				ips: []string{
					"1.1.1.1",
					"fd00::1",
					"1.1.1.1",
				},
			},
			wantIpv4s: []net.IP{
				net.ParseIP("1.1.1.1").To4(),
				net.ParseIP("1.1.1.1").To4(),
			},
			wantIpv6s: []net.IP{
				net.ParseIP("fd00::1"),
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				ips: []string{
					"1.1.1.1",
					"fd00::1",
					"1.1.1.1",
					"1.x.1.1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIpv4s, gotIpv6s, err := ParseIPs(tt.args.ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIpv4s, tt.wantIpv4s) {
				t.Errorf("ParseIPs() gotIpv4s = %v, want %v", gotIpv4s, tt.wantIpv4s)
			}
			if !reflect.DeepEqual(gotIpv6s, tt.wantIpv6s) {
				t.Errorf("ParseIPs() gotIpv6s = %v, want %v", gotIpv6s, tt.wantIpv6s)
			}
		})
	}
}
