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
	"net"
	"testing"
	"testutils"
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

// vim:set sw=4 ts=4 noet:
