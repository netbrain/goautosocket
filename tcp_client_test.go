// Copyright © 2015 Clement 'cmc' Rey <cr.rey.clement@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gas

import (
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

// ----------------------------------------------------------------------------

var (
	server net.Listener
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	var err error

	server, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Println(err)
		os.Exit(666)
	}

	os.Exit(m.Run())
}

// ----------------------------------------------------------------------------

func TestTCPClient_Dial(t *testing.T) {
	c, err := Dial("tcp", server.Addr().String())
	if err != nil {
		t.Error(err)
	}
	if c == nil || c.(*TCPClient).TCPConn == nil {
		t.Error("initialization failed")
	}
	if err := c.Close(); err != nil {
		t.Error(err)
	}
}

func TestTCPClient_DialTCP(t *testing.T) {
	c, err := DialTCP("tcp", nil, server.Addr().(*net.TCPAddr))
	if err != nil {
		t.Error(err)
	}
	if c == nil || c.TCPConn == nil {
		t.Error("initialization failed")
	}
	if err := c.Close(); err != nil {
		t.Error(err)
	}
}

// ----------------------------------------------------------------------------

func TestTCPClient_reconnect(t *testing.T) {
	c, _ := Dial("tcp", server.Addr().String())
	defer c.Close()

	tcpConn1 := c.(*TCPClient).TCPConn
	if err := c.(*TCPClient).reconnect(); err != nil {
		t.Error(err)
	}
	tcpConn2 := c.(*TCPClient).TCPConn
	if tcpConn2 == nil || tcpConn1 == tcpConn2 {
		t.Error("reconnection failed")
	}

	if err := tcpConn1.Close(); err == nil {
		t.Error("tcpConn1 should already be closed")
	} else if err.Error() != "use of closed network connection" {
		t.Error(err)
	}
	if err := tcpConn2.Close(); err != nil {
		t.Error(err)
	}
}
