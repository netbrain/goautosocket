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
