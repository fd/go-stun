// Copyright 2013, Simon Menke. All rights reserved.
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
// author: Simon Menke <simon.menke@gmail.com>

package stun

import (
	"errors"

	"golang.org/x/net/context"
)

// MultiClient will try to discover the NET using multiple stun servers.
type MultiClient struct {
	ServerAddrs  []string // Addresses of the STUN servers
	SoftwareName string   // Name of the Client software (defaults to 'StunClient')
}

// Discover contacts the STUN servers and gets the response of NAT type, host
// for UDP punching
func (mc *MultiClient) Discover(ctx context.Context) (int, *Host, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if mc.SoftwareName == "" {
		mc.SoftwareName = "StunClient"
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(chan multiResponse, len(mc.ServerAddrs))

	for _, addr := range mc.ServerAddrs {
		go mc.discoverSingle(ctx, addr, out)
	}

	var (
		lastResponse multiResponse
		gotResponse  bool
	)

	for i, l := 0, len(mc.ServerAddrs); i < l; i++ {
		select {

		case <-ctx.Done():
			return NAT_ERROR, nil, ctx.Err()

		case res := <-out:
			if res.Err == nil && res.Type != NAT_UNKNOWN && res.Type != NAT_BLOCKED {
				return res.Type, res.Host, res.Err
			}
			lastResponse = res
			gotResponse = true

		}
	}

	if gotResponse {
		return lastResponse.Type, lastResponse.Host, lastResponse.Err
	}

	return NAT_ERROR, nil, errors.New("stun: no response")
}

type multiResponse struct {
	Type int
	Host *Host
	Err  error
}

func (mc *MultiClient) discoverSingle(ctx context.Context, addr string, out chan<- multiResponse) {
	var client = Client{addr, mc.SoftwareName}
	typ, host, err := client.Discover(ctx)
	out <- multiResponse{typ, host, err}
}
