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
	"pythia"
	"reflect"
	"strconv"
	"testing"
	"testutils"
	"testutils/pytest"
)

////////////////////////////////////////////////////////////////////////////////
// Fixture

// QueueFixture contains the common elements for queue tests.
type QueueFixture struct {
	// Queue component
	Queue *Queue

	// Client components
	Clients []*pytest.Conn
}

// Setup an environment for testing the Queue component.
// The queue capacity is configured with capacity.
// A number of clients will be connected to the queue.
func SetupQueueFixture(t *testing.T, capacity int, clients int) *QueueFixture {
	var err error
	f := new(QueueFixture)
	// Setup queue
	t.Log("Setup queue")
	f.Queue = NewQueue()
	addr, err := pythia.LocalAddr()
	if err != nil {
		t.Fatal(err)
	}
	pythia.QueueAddr = addr
	go f.Queue.Run()
	// Setup clients
	t.Log("Setup initial clients")
	f.Clients = make([]*pytest.Conn, clients)
	for i := 0; i < clients; i++ {
		f.Clients[i] = pytest.DialRetry(t, addr)
	}
	return f
}

// TearDown tears down the fixture, closing the connections and shutting down
// the components.
func (f *QueueFixture) TearDown() {
	f.Queue.Shutdown()
	for i := 0; i < len(f.Clients); i++ {
		if f.Clients[i] != nil {
			f.Clients[i].Close()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// Tests

func TestQueueNoop(t *testing.T) {
	testutils.CheckGoroutines(t, func() {
		f := SetupQueueFixture(t, 500, 2)
		f.TearDown()
	})
}

func TestQueueSimple(t *testing.T) {
	f := SetupQueueFixture(t, 500, 2)
	frontend, pool := f.Clients[0], f.Clients[1]
	pool.Send(pythia.Message{
		Message:  pythia.RegisterPoolMsg,
		Capacity: 1,
	})
	task := pytest.ReadTask(t, "hello-world")
	frontend.Send(pythia.Message{
		Message: pythia.LaunchMsg,
		Id:      "test",
		Task:    &task,
		Input:   "Hello world",
	})
	pool.Expect(1, pythia.Message{
		Message: pythia.LaunchMsg,
		Id:      "0:test",
		Task:    &task,
		Input:   "Hello world",
	})
	pool.Send(pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "0:test",
		Status:  pythia.Success,
		Output:  "Hi",
	})
	frontend.Expect(1, pythia.Message{
		Message: pythia.DoneMsg,
		Id:      "test",
		Status:  pythia.Success,
		Output:  "Hi",
	})
	f.TearDown()
}

func TestQueueStatus(t *testing.T) {
	f := SetupQueueFixture(t, 500, 2)
	frontend := f.Clients[0]

	frontend.Send(pythia.Message{
		Message: pythia.StatusMsg,
		Id:      "test",
	})

	// Removing Clients array from Message as this part will differ due to client list
	// referencing the test client as a client but said client is no longer connected
	// when the status is emitted
	status := fillQueueStatus(f.Queue)
	status.Clients = make([]*queueClient, 0)

	msg := <-frontend.Conn.Receive()
	if msg.Message != pythia.DoneMsg || msg.Id != "test" || msg.Status != pythia.Success {
		t.Fatal("Message content mismatching")
	}

	var expected QueueStatus
	expected.Clients = make([]*queueClient, 0)
	json.Unmarshal([]byte(msg.Output), &expected)

	// The content of the Waiting list is not compared because it's not efficient
	// and it's not really interesting because the list is supposed to be empty
	if expected.Capacity != status.Capacity ||
		expected.Available != status.Available ||
		!reflect.DeepEqual(expected.Jobs, status.Jobs) ||
		!reflect.DeepEqual(expected.Waiting.Len(), status.Waiting.Len()) ||
		!expected.CreationDate.Equal(status.CreationDate) {

		t.Error("Capacity : " + strconv.FormatBool(expected.Capacity != status.Capacity))
		t.Error("Available : " + strconv.FormatBool(expected.Available != status.Available))
		t.Error("Jobs : " + strconv.FormatBool(!reflect.DeepEqual(expected.Jobs, status.Jobs)))
		t.Error("Waiting : " + strconv.FormatBool(!reflect.DeepEqual(expected.Waiting, status.Waiting)))
		t.Error("CreationDate : " + strconv.FormatBool(!expected.CreationDate.Equal(status.CreationDate)))
	}

	f.TearDown()
}

// vim:set sw=4 ts=4 noet:
