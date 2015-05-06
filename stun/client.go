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

package stun

import (
	"golang.org/x/net/context"
)

// Client holds all STUN server details.
type Client struct {
	ServerAddr   string // Address of the STUN server
	SoftwareName string // Name of the Client software (defaults to 'StunClient')
}

// DefaultClient for global methods
var DefaultClient = Client{
	ServerAddr: "stun1.voiceeclipse.net:3478",
}

// Discover contacts the STUN server and gets the response of NAT type, host
// for UDP punching
func Discover(ctx context.Context) (int, *Host, error) {
	return DefaultClient.Discover(ctx)
}

// Discover contacts the STUN server and gets the response of NAT type, host
// for UDP punching
func (client *Client) Discover(ctx context.Context) (int, *Host, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return client.discover(ctx)
}
