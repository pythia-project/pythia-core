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
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const (
	// Duration between two keep-alive messages sent.
	keepAlivePeriod = 30 * time.Second

	// Duration after which the connection is considered closed if no message
	// has been received.
	readTimeout = 3 * keepAlivePeriod
)

// Error to return if the connection was closed
var closedError = errors.New("Connection closed")

// MessageResult is an auxiliary structure for passing messages to the writer
// goroutine.
type messageResult struct {
	Msg    Message      // what to send
	Result chan<- error // where to write the result of the operation
}

// Conn is a wrapper over net.Conn, reading and writing Messages.
type Conn struct {
	// The underlying connection.
	conn net.Conn

	// Incoming messages channel.
	input chan Message

	// Outgoing messages channel.
	output chan messageResult

	// Channel to ask reader and writer goroutines to quit.
	quit chan bool

	// Flag to ignore errors after closing the connection.
	closed bool
}

// WrapConn wraps a stream network connection into a Message-oriented
// connection. The raw conn connection shall not be used by the user anymore.
func WrapConn(conn net.Conn) *Conn {
	c := new(Conn)
	c.conn = conn
	c.input = make(chan Message)
	c.output = make(chan messageResult)
	c.quit = make(chan bool, 1)
	go c.reader()
	go c.writer()
	return c
}

// The reader goroutine parses the Messages and put them in the input channel.
// Keep-alive messages are discarded.
func (c *Conn) reader() {
	defer close(c.input)
	dec := json.NewDecoder(c.conn)
	for {
		c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		var msg Message
		err := dec.Decode(&msg)
		if c.closed {
			return
		} else if err == io.EOF {
			log.Println("Connection closed on remote side.")
			c.Close()
			return
		} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			log.Println("Connection timed out.")
			c.Close()
			return
		} else if err != nil {
			log.Print(err)
		} else if msg.Message != KeepAliveMsg {
			c.input <- msg
		}
	}
}

// The writer goroutine sends Messages and keep-alives
func (c *Conn) writer() {
	keepAliveTicker := time.NewTicker(keepAlivePeriod)
	defer keepAliveTicker.Stop()
	sendKeepAlive := true
	enc := json.NewEncoder(c.conn)
	for {
		select {
		case mr := <-c.output:
			msg, result := mr.Msg, mr.Result
			result <- enc.Encode(msg)
			sendKeepAlive = false
		case <-keepAliveTicker.C:
			if sendKeepAlive {
				err := enc.Encode(Message{Message: KeepAliveMsg})
				if err != nil {
					log.Println("Error sending keep-alive message:", err)
				}
			}
			sendKeepAlive = true
		case <-c.quit:
			c.sendQuit()
			return
		}
	}
}

// Dial connects to the address addr and returns a Message-oriented connection.
func Dial(addr net.Addr) (*Conn, error) {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return WrapConn(conn), nil
}

// Receive returns the channel from which incoming messages can be retrieved.
// The channel is closed when the connection is closed.
func (c *Conn) Receive() <-chan Message {
	return c.input
}

// Send sends a message through the connection.
func (c *Conn) Send(msg Message) error {
	if c.closed {
		return closedError
	}
	result := make(chan error)
	select {
	case c.output <- messageResult{Msg: msg, Result: result}:
	case <-c.quit:
		c.sendQuit()
		return closedError
	}
	select {
	case err := <-result:
		return err
	case <-c.quit:
		c.sendQuit()
		return closedError
	}
}

// sendQuit signals the quit channel, but does not block. This requires the
// quit channel to be buffered.
func (c *Conn) sendQuit() {
	select {
	case c.quit <- true:
	default:
	}
}

// Close closes the connection. The receive channel will also be closed.
// Further sends will cause errors.
func (c *Conn) Close() error {
	c.sendQuit()
	c.closed = true
	return c.conn.Close()
}

// vim:set sw=4 ts=4 noet:
