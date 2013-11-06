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
	"fmt"
	"pythia"
	"testing"
	"testutils"
	"testutils/pytest"
)

////////////////////////////////////////////////////////////////////////////////
// Fixture

// PoolFixture contains the common elements for pool tests.
type PoolFixture struct {
	// Mock queue component
	Queue *pythia.Listener

	// Pool component
	Pool *Pool

	// Queue->Pool connection
	Conn *pytest.Conn
}

// Setup an environment for testing the Pool component.
// The pool capacity is configured with capacity. The test will fail if the pool
// does not register correctly.
func SetupPoolFixture(t *testing.T, capacity int) *PoolFixture {
	var err error
	f := new(PoolFixture)
	// Setup mock queue
	t.Log("Setup queue")
	addr, err := pythia.LocalAddr()
	if err != nil {
		t.Fatal(err)
	}
	f.Queue, err = pythia.Listen(addr)
	if err != nil {
		t.Fatal(err)
	}
	// Setup pool
	t.Log("Setup pool")
	f.Pool = NewPool()
	f.Pool.QueueAddr = addr
	f.Pool.Capacity = capacity
	f.Pool.UmlPath = pytest.UmlPath
	f.Pool.EnvDir = pytest.VmDir
	f.Pool.TasksDir = pytest.TasksDir
	go f.Pool.Run()
	// Establish connection
	t.Log("Establish connection")
	conn, err := f.Queue.Accept()
	if err != nil {
		f.Queue.Close()
		t.Fatal(err)
	}
	f.Conn = &pytest.Conn{t, conn}
	// Wait for register-pool message
	f.Conn.Expect(2, pythia.Message{
		Message:  pythia.RegisterPoolMsg,
		Capacity: capacity,
	})
	return f
}

// TearDown tears down the fixture, closing the connections and shutting down
// the components.
func (f *PoolFixture) TearDown() {
	f.Conn.Close()
	f.Queue.Close()
}

////////////////////////////////////////////////////////////////////////////////
// Tests

func TestPoolNoop(t *testing.T) {
	testutils.CheckGoroutines(t, func() {
		f := SetupPoolFixture(t, 1)
		f.TearDown()
	})
}

func TestPoolHelloWorld(t *testing.T) {
	task := pytest.ReadTask(t, "hello-world")
	f := SetupPoolFixture(t, 1)
	f.Conn.Send(pythia.Message{
		Message: pythia.LaunchMsg,
		Id:      "hello",
		Task:    &task,
		Input:   "",
	})
	f.Conn.Expect(5, pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "hello",
		Status:  pythia.Success,
		Output:  "Hello world!\n",
	})
	f.TearDown()
}

func TestPoolExceedCapacity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	task := pytest.ReadTask(t, "timeout")
	task.Limits.Time = 5
	f := SetupPoolFixture(t, 2)
	for i := 1; i <= 3; i++ {
		f.Conn.Send(pythia.Message{
			Message: pythia.LaunchMsg,
			Id:      fmt.Sprint(i),
			Task:    &task,
		})
	}
	f.Conn.Expect(2, pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "3",
		Status:  pythia.Error,
		Output:  "Pool capacity exceeded",
	})
	f.Conn.Expect(6, pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "1",
		Status:  pythia.Timeout,
		Output:  "Start\n",
	}, pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "2",
		Status:  pythia.Timeout,
		Output:  "Start\n",
	})
	f.TearDown()
}

// vim:set sw=4 ts=4 noet:
