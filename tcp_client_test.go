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
	c.Close()
}

func TestTCPClient_DialTCP(t *testing.T) {
	c, err := DialTCP("tcp", nil, server.Addr().(*net.TCPAddr))
	if err != nil {
		t.Error(err)
	}
	if c == nil || c.TCPConn == nil {
		t.Error("initialization failed")
	}
	c.Close()
}
