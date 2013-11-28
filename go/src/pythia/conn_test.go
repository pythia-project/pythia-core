// Copyright 2013 The Pythia Authors.
// This file is part of Pythia.
//
// Pythia is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// Pythia is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

package pythia

import (
	"encoding/json"
	"net"
	"testing"
	"testutils"
	"time"
)

// Flush receive channel of c, generating errors for messages left.
func connTestFlush(t *testing.T, c *Conn) {
	for m := range c.Receive() {
		t.Errorf("Unexpected message: `%v`.", m)
	}
}

// Test cleanup when closing the connection on our side.
func TestConnCloseLocal(t *testing.T) {
	testutils.CheckGoroutines(t, func() {
		local, _ := net.Pipe()
		conn := WrapConn(local)
		conn.Close()
	})
}

// Test cleanup when remote closes the connection.
func TestConnCloseRemote(t *testing.T) {
	testutils.CheckGoroutines(t, func() {
		local, remote := net.Pipe()
		WrapConn(local)
		remote.Close()
	})
}

// Test a simple message transfer
func TestConnSimpleMessage(t *testing.T) {
	raw1, raw2 := net.Pipe()
	c1, c2 := WrapConn(raw1), WrapConn(raw2)
	msg := Message{Message: DoneMsg, Id: "1"}
	go c1.Send(msg)
	received := <-c2.Receive()
	testutils.Expect(t, "received", msg, received)
	c1.Close()
	c2.Close()
	connTestFlush(t, c2)
}

// Check that a connection sends a keep-alive message
func TestConnKeepAliveSend(t *testing.T) {
	KeepAliveInterval = 100 * time.Millisecond
	raw1, raw2 := net.Pipe()
	c1 := WrapConn(raw1)
	defer c1.Close()
	dec := json.NewDecoder(raw2)
	start := time.Now()
	var msg Message
	dec.Decode(&msg)
	elapsed := time.Since(start)
	t.Log("Received", msg)
	if msg.Message != KeepAliveMsg {
		t.Error("Message is not a keep-alive message")
	}
	if elapsed.Seconds() < 0.09 {
		t.Error("Message arrived too early:", elapsed)
	} else if elapsed.Seconds() > 0.11 {
		t.Error("Message arrived too late:", elapsed)
	}
}

// Check that a connection closes when no keep-alive messages are sent
func TestConnKeepAliveClose(t *testing.T) {
	KeepAliveInterval = 10 * time.Millisecond
	addr, err := LocalAddr()
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			if _, err := l.Accept(); err == nil {
				t.Log("New connection")
			} else {
				break
			}
		}
	}()
	testutils.CheckGoroutines(t, func() {
		conn, err := Dial(addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		timer := time.NewTimer(100 * time.Millisecond)
		defer timer.Stop()
		select {
		case msg, ok := <-conn.Receive():
			if ok {
				t.Error("Unexpected message", msg)
			}
		case <-timer.C:
			t.Error("Timeout")
		}
	})
}

// Test DialRetry
func TestConnDialRetry(t *testing.T) {
	addr, err := LocalAddr()
	if err != nil {
		t.Fatal(err)
	}
	c := make(chan bool)
	go func() {
		conn := DialRetry(addr)
		conn.Close()
		c <- true
	}()
	time.Sleep(50 * time.Millisecond)
	l, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	l.Accept()
	<-c
}

// vim:set sw=4 ts=4 noet:
