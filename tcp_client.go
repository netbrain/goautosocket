// Copyright © 2015 Clement 'cmc' Rey <cr.rey.clement@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gas

import "net"

// ----------------------------------------------------------------------------

// TCPClient provides a TCP connection with auto-reconnect capabilities.
//
// It embeds a *net.TCPConn and thus implements the net.Conn interface.
type TCPClient struct {
	*net.TCPConn
}

// Dial returns a new *TCPClient.
//
// The new client connects to the remote address `raddr` on the network `network`,
// which must be "tcp", "tcp4", or "tcp6".
//
// This complements net package's Dial function.
func Dial(network, addr string) (*TCPClient, error) {
	raddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	return DialTCP(network, nil, raddr)
}

// DialTCP returns a new *TCPClient.
//
// The new client connects to the remote address `raddr` on the network `network`,
// which must be "tcp", "tcp4", or "tcp6".
// If `laddr` is not nil, it is used as the local address for the connection.
//
// This overrides net.TCPConn's DialTCP function.
func DialTCP(network string, laddr, raddr *net.TCPAddr) (*TCPClient, error) {
	conn, err := net.DialTCP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}

	return &TCPClient{conn}, nil
}
