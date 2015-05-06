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
	"errors"
	"net"
	"time"

	"golang.org/x/net/context"
)

// padding the length of the byte slice to multiple of 4
func padding(b []byte) []byte {
	l := uint16(len(b))
	return append(b, make([]byte, align(l)-l)...)
}

// align the uint16 number to the smallest multiple of 4, which is larger than
// or equal to the uint16 number
func align(l uint16) uint16 {
	return (l + 3) & 0xfffc
}

func dialWithContext(ctx context.Context, addr string) (net.Conn, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return net.Dial("udp", addr)
	}

	conn, err := net.DialTimeout("udp", addr, deadline.Sub(time.Now()))
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(deadline)

	return conn, nil
}

func (client *Client) sendBindingReq(ctx context.Context, destAddr string) (*packet, string, error) {
	connection, err := dialWithContext(ctx, destAddr)
	if err != nil {
		return nil, "", err
	}

	defer connection.Close()

	packet := newPacket()
	packet.types = type_BINDING_REQUEST
	attribute := newSoftwareAttribute(packet, client.SoftwareName)
	packet.addAttribute(*attribute)
	attribute = newFingerprintAttribute(packet)
	packet.addAttribute(*attribute)

	localAddr := connection.LocalAddr().String()
	packet, err = packet.send(connection)
	if err != nil {
		return nil, "", err
	}

	return packet, localAddr, err
}

func (client *Client) sendChangeReq(ctx context.Context, changeIP bool, changePort bool) (*packet, error) {
	connection, err := dialWithContext(ctx, client.ServerAddr)
	if err != nil {
		return nil, err
	}

	defer connection.Close()

	// construct packet
	packet := newPacket()
	packet.types = type_BINDING_REQUEST
	attribute := newSoftwareAttribute(packet, client.SoftwareName)
	packet.addAttribute(*attribute)
	attribute = newChangeReqAttribute(packet, changeIP, changePort)
	packet.addAttribute(*attribute)
	attribute = newFingerprintAttribute(packet)
	packet.addAttribute(*attribute)

	packet, err = packet.send(connection)
	if err != nil {
		return nil, err
	}

	return packet, err
}

func (client *Client) test1(ctx context.Context, destAddr string) (*packet, string, bool, *Host, error) {
	packet, localAddr, err := client.sendBindingReq(ctx, destAddr)
	if err != nil {
		return nil, "", false, nil, err
	}
	if packet == nil {
		return nil, "", false, nil, nil
	}

	hm := packet.xorMappedAddr()
	// rfc 3489 doesn't require the server return xor mapped address
	if hm == nil {
		hm = packet.mappedAddr()
		if hm == nil {
			return nil, "", false, nil, errors.New("No mapped address")
		}
	}

	hc := packet.changedAddr()
	if hc == nil {
		return nil, "", false, nil, errors.New("No changed address")
	}
	changeAddr := hc.TransportAddr()
	identical := localAddr == hm.TransportAddr()

	return packet, changeAddr, identical, hm, nil
}

func (client *Client) test2(ctx context.Context) (*packet, error) {
	return client.sendChangeReq(ctx, true, true)
}

func (client *Client) test3(ctx context.Context) (*packet, error) {
	return client.sendChangeReq(ctx, false, true)
}

// follow rfc 3489 and 5389
func (client *Client) discover(ctx context.Context) (int, *Host, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	packet, changeAddr, identical, host, err := client.test1(ctx, client.ServerAddr)
	if err != nil {
		return NAT_ERROR, nil, err
	}
	if packet == nil {
		return NAT_BLOCKED, nil, err
	}

	// detect symetric
	if identical {
		packet, err = client.test2(ctx)
		if err != nil {
			return NAT_ERROR, host, err
		}
		if packet != nil {
			return NAT_NONE, host, nil
		}
		return NAT_SYMETRIC_UDP_FIREWALL, host, nil
	}

	// detect full nat
	packet, err = client.test2(ctx)
	if err != nil {
		return NAT_ERROR, host, err
	}
	if packet != nil {
		return NAT_FULL, host, nil
	}

	packet, _, identical, _, err = client.test1(ctx, changeAddr)
	if err != nil {
		return NAT_ERROR, host, err
	}
	if packet == nil {
		// It should be NAT_BLOCKED, but will be
		// detected in the first step. So this will
		// never happen.
		return NAT_UNKNOWN, host, nil
	}
	if identical {
		packet, err = client.test3(ctx)
		if err != nil {
			return NAT_ERROR, host, err
		}
		if packet == nil {
			return NAT_PORT_RESTRICTED, host, nil
		}
		return NAT_RESTRICTED, host, nil
	}

	return NAT_SYMETRIC, host, nil
}
