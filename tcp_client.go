// Copyright © 2015 Clement 'cmc' Rey <cr.rey.clement@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gas

import (
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

// ----------------------------------------------------------------------------

// TCPClient provides a TCP connection with auto-reconnect capabilities.
//
// It embeds a *net.TCPConn and thus implements the net.Conn interface.
//
// TCPClient can be safely used from multiple goroutines.
type TCPClient struct {
	*net.TCPConn

	lock sync.RWMutex
}

// Dial returns a new net.Conn.
//
// The new client connects to the remote address `raddr` on the network `network`,
// which must be "tcp", "tcp4", or "tcp6".
//
// This complements net package's Dial function.
func Dial(network, addr string) (net.Conn, error) {
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

	return &TCPClient{TCPConn: conn, lock: sync.RWMutex{}}, nil
}

// ----------------------------------------------------------------------------

// reconnect builds a new TCP connection to replace the embedded *net.TCPConn.
//
// TODO: keep old socket configuration (timeout, linger...).
func (c *TCPClient) reconnect() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	raddr := c.TCPConn.RemoteAddr()
	conn, err := net.DialTCP(raddr.Network(), nil, raddr.(*net.TCPAddr))
	if err != nil {
		return err
	}

	c.TCPConn.Close()
	c.TCPConn = conn
	return nil
}

// ----------------------------------------------------------------------------

// Read wraps net.TCPConn's Read method with reconnect capabilities.
func (c *TCPClient) Read(b []byte) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 10
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.Read(b)
		if err == nil {
			return n, err
		} else if err.Error() == "EOF" {
			c.lock.RUnlock()
			if c.reconnect() != nil {
				time.Sleep(t)
			}
			c.lock.RLock()
		} else {
			return n, err
		}
		t *= 2
	}

	return -1, ErrMaxRetries
}

// ReadFrom wraps net.TCPConn's Read method with reconnect capabilities.
func (c *TCPClient) ReadFrom(r io.Reader) (int64, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 10
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.ReadFrom(r)
		if err == nil {
			return n, err
		} else if err.Error() == "EOF" {
			c.lock.RUnlock()
			if c.reconnect() != nil {
				time.Sleep(t)
			}
			c.lock.RLock()
		} else {
			return n, err
		}
		t *= 2
	}

	return -1, ErrMaxRetries
}

// Write wraps net.TCPConn's Read method with reconnect capabilities.
func (c *TCPClient) Write(b []byte) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 10
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.Write(b)
		if err == nil {
			return n, err
		} else {
			switch e := err.(type) {
			case *net.OpError:
				if e.Err.(syscall.Errno) == 0x20 ||
					e.Err.(syscall.Errno) == 0x68 {
					c.lock.RUnlock()
					if c.reconnect() != nil {
						time.Sleep(t)
					}
					c.lock.RLock()
				} else {
					return n, err
				}
			default:
				return n, err
			}
		}
		t *= 2
	}

	return -1, ErrMaxRetries
}
