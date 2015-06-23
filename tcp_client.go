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
// Use the SetMaxRetries() and SetRetryInterval() methods to configure retry
// values; otherwise they default to maxRetries=5 and retryInterval=100ms.
//
// TCPClient can be safely used from multiple goroutines.
type TCPClient struct {
	*net.TCPConn

	lock sync.RWMutex

	maxRetries    int
	retryInterval time.Duration
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

	return &TCPClient{
		TCPConn: conn,

		lock: sync.RWMutex{},

		maxRetries:    5,
		retryInterval: 100 * time.Millisecond,
	}, nil
}

// ----------------------------------------------------------------------------

// SetMaxRetries sets the retry limit for the TCPClient.
//
// Assuming i is the current retry iteration, the total sleep time is
// t = retryInterval * (2^i)
//
// This function completely Lock()s the TCPClient.
func (c *TCPClient) SetMaxRetries(maxRetries int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.maxRetries = maxTries
}

// GetMaxRetries gets the retry limit for the TCPClient.
//
// Assuming i is the current retry iteration, the total sleep time is
// t = retryInterval * (2^i)
func (c *TCPClient) GetMaxRetries(maxRetries int) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.maxRetries
}

// SetRetryInterval sets the retry interval for the TCPClient.
//
// Assuming i is the current retry iteration, the total sleep time is
// t = retryInterval * (2^i)
//
// This function completely Lock()s the TCPClient.
func (c *TCPClient) SetRetryInterval(retryInterval time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.retryInterval = retryInterval
}

// GetRetryInterval gets the retry interval for the TCPClient.
//
// Assuming i is the current retry iteration, the total sleep time is
// t = retryInterval * (2^i)
func (c *TCPClient) GetRetryInterval(retryInterval time.Duration) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.retryInterval
}

// ----------------------------------------------------------------------------

// reconnect builds a new TCP connection to replace the embedded *net.TCPConn.
//
// This function completely Lock()s the TCPClient.
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
//
// It will return ErrMaxRetries if the retry limit is reached.
func (c *TCPClient) Read(b []byte) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 5
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.Read(b)
		if err == nil {
			return n, err
		}
		switch e := err.(type) {
		case *net.OpError:
			if e.Err.(syscall.Errno) == syscall.EPIPE ||
				e.Err.(syscall.Errno) == syscall.ECONNRESET {
				c.lock.RUnlock()
				if c.reconnect() != nil {
					time.Sleep(t)
				}
				c.lock.RLock()
			}
		default:
			if err.Error() == "EOF" {
				c.lock.RUnlock()
				if c.reconnect() != nil {
					time.Sleep(t)
				}
				c.lock.RLock()
			} else {
				return n, err
			}
		}
		t *= 2
	}

	return -1, ErrMaxRetries
}

// ReadFrom wraps net.TCPConn's Read method with reconnect capabilities.
//
// It will return ErrMaxRetries if the retry limit is reached.
func (c *TCPClient) ReadFrom(r io.Reader) (int64, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 5
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.ReadFrom(r)
		if err == nil {
			return n, err
		}
		switch e := err.(type) {
		case *net.OpError:
			if e.Err.(syscall.Errno) == syscall.EPIPE ||
				e.Err.(syscall.Errno) == syscall.ECONNRESET {
				c.lock.RUnlock()
				if c.reconnect() != nil {
					time.Sleep(t)
				}
				c.lock.RLock()
			}
		default:
			if err.Error() == "EOF" {
				c.lock.RUnlock()
				if c.reconnect() != nil {
					time.Sleep(t)
				}
				c.lock.RLock()
			} else {
				return n, err
			}
		}
		t *= 2
	}

	return -1, ErrMaxRetries
}

// Write wraps net.TCPConn's Read method with reconnect capabilities.
//
// It will return ErrMaxRetries if the retry limit is reached.
func (c *TCPClient) Write(b []byte) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	maxTries := 5
	t := time.Millisecond * 100

	for i := 0; i < maxTries; i++ {
		n, err := c.TCPConn.Write(b)
		if err == nil {
			return n, err
		}
		switch e := err.(type) {
		case *net.OpError:
			if e.Err.(syscall.Errno) == syscall.EPIPE ||
				e.Err.(syscall.Errno) == syscall.ECONNRESET {
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
		t *= 2
	}

	return -1, ErrMaxRetries
}
