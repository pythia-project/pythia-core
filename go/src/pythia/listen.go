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
)

// Listener is a wrapper around net.Listener providing a Message-oriented
// interface.
type Listener struct {
	// The address on which the listener listens
	Addr net.Addr

	// The underlying listener
	listener net.Listener
}

// Listen announces on address addr and listens for connections.
func Listen(addr net.Addr) (*Listener, error) {
	listener, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	l := new(Listener)
	l.Addr = addr
	l.listener = listener
	return l, err
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (*Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return WrapConn(conn), nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return l.listener.Close()
}

// vim:set sw=4 ts=4 noet:
