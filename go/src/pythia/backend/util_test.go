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

package backend

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"pythia"
	"reflect"
	"testing"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// Common utility functions for tests

// Environment variables exported from make
var (
	TopDir   = os.Getenv("TOP_DIR")
	OutDir   = path.Join(TopDir, os.Getenv("OUT_DIR"))
	UmlPath  = path.Join(TopDir, os.Getenv("UML"))
	VmDir    = path.Join(TopDir, os.Getenv("VM_OUT_DIR"))
	TasksDir = path.Join(TopDir, os.Getenv("TASKS_OUT_DIR"))
)

// ReadTask reads a task description from a file.
func ReadTask(basename string) (*pythia.Task, error) {
	content, err := ioutil.ReadFile(path.Join(TasksDir, basename+".task"))
	if err != nil {
		return nil, err
	}
	task := new(pythia.Task)
	if json.Unmarshal(content, task) != nil {
		return nil, err
	}
	return task, nil
}

////////////////////////////////////////////////////////////////////////////////
// Fixtures

// ConnFixture is the base fixture for connection-based tests.
type ConnFixture struct {
	T    *testing.T
	Conn *pythia.Conn
}

// Send a message through the connection.
func (f *ConnFixture) Send(msg pythia.Message) {
	f.T.Log(">>", msg)
	f.Conn.Send(msg)
}

// Read len(expected) messages from the connection. Fail if a message was not in
// the expected set (order does not matter) or an expected message was not
// received. If nothing is received for more than timeout seconds, fail and
// continue.
func (f *ConnFixture) Expect(timeout int, expected ...pythia.Message) {
	ok := make([]bool, len(expected))
	for i := 0; i < len(expected); i++ {
		select {
		case msg := <-f.Conn.Receive():
			found := false
			for i, m := range expected {
				if !ok[i] && reflect.DeepEqual(msg, m) {
					f.T.Log("<<", msg)
					ok[i] = true
					found = true
					break
				}
			}
			if !found {
				f.T.Error("<<(unexpected)", msg)
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			f.T.Error("<< timed out")
			break
		}
	}
	for i, msg := range expected {
		if !ok[i] {
			f.T.Error("<<(missing)", msg)
		}
	}
}

// Check for an unexpected message waiting in the input channel and close the
// connection.
func (f *ConnFixture) TearDown() {
	select {
	case msg := <-f.Conn.Receive():
		f.T.Error("<<(unexpected)", msg)
	default:
	}
	f.Conn.Close()
}

// vim:set sw=4 ts=4 noet:
