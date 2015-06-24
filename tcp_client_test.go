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
	"sync"
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
	defer server.Close()

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

// ----------------------------------------------------------------------------

func TestTCPClient_Write(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped (short mode)")
	}

	// open a server socket
	s, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Error(err)
	}
	// save the original port
	addr := s.Addr()

	nbClients := 10
	// connect nbClients clients to the server
	clients := make([]net.Conn, nbClients)
	for i := 0; i < len(clients); i++ {
		c, err := Dial("tcp", s.Addr().String())
		if err != nil {
			t.Error(err)
		}
		c.(*TCPClient).SetMaxRetries(10)
		c.(*TCPClient).SetRetryInterval(10 * time.Millisecond)
		defer c.Close()
		clients[i] = c
	}

	// shut down and boot up the server randomly
	var swg sync.WaitGroup
	swg.Add(1)
	go func() {
		defer swg.Done()
		for i := 0; i < 10; i++ {
			log.Println("server up")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(30)))
			if err := s.Close(); err != nil {
				t.Error(err)
			}
			log.Println("server down")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(10)))
			s, err = net.Listen("tcp", addr.String())
			if err != nil {
				t.Error(err)
			}
		}
	}()

	// clients concurrently writes to the server
	var cwg sync.WaitGroup
	for i, c := range clients {
		cwg.Add(1)
		go func(ii int, cc net.Conn) {
			str := []byte("hello, world!")
			defer cwg.Done()
			for {
				if _, err := cc.Write(str); err != nil {
					switch e := err.(type) {
					case Error:
						if e == ErrMaxRetries {
							log.Println("client", ii, "leaving, reached retry limit while writing")
							return
						}
					default:
						t.Error(err)
					}
				}
			}
		}(i, c)
	}

	// terminates the server indefinitely
	swg.Wait()
	if err := s.Close(); err != nil {
		t.Error(err)
	}

	// wait for clients to give up
	cwg.Wait()
}

func TestTCPClient_Read(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped (short mode)")
	}

	// open a server socket
	s, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Error(err)
	}
	// save the original port
	addr := s.Addr()

	nbClients := 10
	// connect nbClients clients to the server
	clients := make([]net.Conn, nbClients)
	for i := 0; i < len(clients); i++ {
		c, err := Dial("tcp", s.Addr().String())
		if err != nil {
			t.Error(err)
		}
		c.(*TCPClient).SetMaxRetries(5)
		c.(*TCPClient).SetRetryInterval(10 * time.Millisecond)
		defer c.Close()
		clients[i] = c
	}

	// shut down and boot up the server randomly
	var swg sync.WaitGroup
	swg.Add(1)
	go func() {
		defer swg.Done()
		for i := 0; i < 10; i++ {
			log.Println("server up")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(30)))
			if err := s.Close(); err != nil {
				t.Error(err)
			}
			log.Println("server down")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(10)))
			s, err = net.Listen("tcp", addr.String())
			if err != nil {
				t.Error(err)
			}
		}
	}()

	// clients concurrently reads from the server
	var cwg sync.WaitGroup
	for i, c := range clients {
		cwg.Add(1)
		go func(ii int, cc net.Conn) {
			str := []byte("hello, world!")
			b := make([]byte, len(str))
			defer cwg.Done()
			for {
				if _, err := cc.Read(b); err != nil {
					switch e := err.(type) {
					case Error:
						if e == ErrMaxRetries {
							log.Println("client", ii, "leaving, reached retry limit while reading")
							return
						}
					default:
						t.Error(err)
					}
				}
			}
		}(i, c)
	}

	// terminates the server indefinitely
	swg.Wait()
	if err := s.Close(); err != nil {
		t.Error(err)
	}

	// wait for clients to give up
	cwg.Wait()
}

// ----------------------------------------------------------------------------

func ExampleTCPClient() {
	// open a server socket
	s, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	// save the original port
	addr := s.Addr()

	// connect a client to the server
	c, err := Dial("tcp", s.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// shut down and boot up the server randomly
	var swg sync.WaitGroup
	swg.Add(1)
	go func() {
		defer swg.Done()
		for i := 0; i < 5; i++ {
			log.Println("server up")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(20)))
			if err := s.Close(); err != nil {
				log.Fatal(err)
			}
			log.Println("server down")
			time.Sleep(time.Millisecond * 100 * time.Duration(rand.Intn(20)))
			s, err = net.Listen("tcp", addr.String())
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	// client writes to the server and reconnects when it has to
	// this is the interesting part
	var cwg sync.WaitGroup
	cwg.Add(1)
	go func() {
		defer cwg.Done()
		for {
			if _, err := c.Write([]byte("hello, world!\n")); err != nil {
				switch e := err.(type) {
				case Error:
					if e == ErrMaxRetries {
						log.Println("client leaving, reached retry limit")
						return
					}
				default:
					log.Fatal(err)
				}
			}
		}
	}()

	// terminates the server indefinitely
	swg.Wait()
	if err := s.Close(); err != nil {
		log.Fatal(err)
	}

	// wait for the client to give up
	cwg.Wait()
}
