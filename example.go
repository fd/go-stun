// Copyright 2013, Cong Ding. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// author: Cong Ding <dinggnu@gmail.com>

package main

import (
	"fmt"
	"time"

	"github.com/fd/go-stun/stun"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mc := stun.MultiClient{
		ServerAddrs: []string{
			"stun.l.google.com:19302",
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
			"stun01.sipphone.com:3478",
			"stun.ekiga.net:3478",
			"stun.fwdnet.net:3478",
			"stun.ideasip.com:3478",
			"stun.iptel.org:3478",
			"stun.rixtelecom.se:3478",
			"stun.schlund.de:3478",
			"stunserver.org:3478",
			"stun.softjoys.com:3478",
			"stun.voiparound.com:3478",
			"stun.voipbuster.com:3478",
			"stun.voipstunt.com:3478",
			"stun.voxgratia.org:3478",
			"stun.xten.com:3478",
		},
	}

	nat, host, err := mc.Discover(ctx)
	if err != nil {
		fmt.Println(err)
	}

	switch nat {
	case stun.NAT_ERROR:
		fmt.Println("Test failed")
	case stun.NAT_UNKNOWN:
		fmt.Println("Unexpected response from the STUN server")
	case stun.NAT_BLOCKED:
		fmt.Println("UDP is blocked")
	case stun.NAT_FULL:
		fmt.Println("Full cone NAT")
	case stun.NAT_SYMETRIC:
		fmt.Println("Symetric NAT")
	case stun.NAT_RESTRICTED:
		fmt.Println("Restricted NAT")
	case stun.NAT_PORT_RESTRICTED:
		fmt.Println("Port restricted NAT")
	case stun.NAT_NONE:
		fmt.Println("Not behind a NAT")
	case stun.NAT_SYMETRIC_UDP_FIREWALL:
		fmt.Println("Symetric UDP firewall")
	}

	if host != nil {
		fmt.Println(host.Family())
		fmt.Println(host.Ip())
		fmt.Println(host.Port())
	}
}
