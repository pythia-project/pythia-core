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
	"flag"
	"log"
	"net"
	"os"
	"pythia"
)

// A Pool is a component that launches sandboxes on the local machine.
// Each Pool has a limit on the number of sandboxes that can run concurrently.
//
// A Pool connects to the Queue, advertises its limits, and waits for jobs to
// execute.
type Pool struct {
	// The address of the Queue
	QueueAddr net.Addr

	// Maximum number of sandboxes that may run at the same time
	Capacity int

	// Path to the UML executable
	UmlPath string

	// Path to the directory containing the environments
	EnvDir string

	// Path to the directory containing the tasks
	TasksDir string

	// Connection to the queue
	conn *pythia.Conn
}

// Setup the parameters with the command line flags in args.
func (pool *Pool) Setup(args []string) {
	fs := flag.NewFlagSet(os.Args[0]+" pool", flag.ExitOnError)
	queue := fs.String("queue", "127.0.0.1:9000", "queue address")
	fs.IntVar(&pool.Capacity, "capacity", 1, "max parallel sandboxes")
	fs.StringVar(&pool.UmlPath, "uml", "vm/uml", "path to the UML executable")
	fs.StringVar(&pool.EnvDir, "envdir", "vm", "environments directory")
	fs.StringVar(&pool.TasksDir, "tasksdir", "tasks", "tasks directory")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	queueaddr, err := pythia.ParseAddr(*queue)
	if err != nil {
		log.Fatal(err)
	}
	pool.QueueAddr = queueaddr
}

// Run the Pool component.
func (pool *Pool) Run() {
	conn, err := pythia.Dial(pool.QueueAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	pool.conn = conn
	// Tokens is a buffered channel to enforce the capacity. Values do not
	// matter.
	tokens := make(chan bool, pool.Capacity)
	for i := 0; i < pool.Capacity; i++ {
		tokens <- true
	}
	conn.Send(pythia.Message{
		Message:  pythia.RegisterPoolMsg,
		Capacity: pool.Capacity,
	})
	for msg := range conn.Receive() {
		switch msg.Message {
		case pythia.LaunchMsg:
			select {
			case <-tokens:
				go func(msg pythia.Message) {
					pool.doJob(msg.Id, msg.Task, msg.Input)
					tokens <- true
				}(msg)
			default:
				log.Println("Capacity exceeded, cannot handle job.")
				conn.Send(pythia.Message{
					Message: pythia.DoneMsg,
					Id:      msg.Id,
					Status:  pythia.Error,
					Output:  "Pool capacity exceeded",
				})
			}
		default:
			log.Println("Ignoring message", msg.Message)
		}
	}
	// BUG(vianney): Abort remaining jobs in Pool after the connection has
	// closed.
	// BUG(vianney): Reconnect Pool to Queue automatically.
}

// DoJob executes a job and sends the result to the queue.
// This function is meant to be run in its own goroutine, as it will block
// until the end of the job execution.
func (pool *Pool) doJob(id string, task *pythia.Task, input string) {
	job := NewJob()
	job.Task = *task
	job.Input = input
	job.UmlPath = pool.UmlPath
	job.EnvDir = pool.EnvDir
	job.TasksDir = pool.TasksDir
	status, output := job.Execute()
	pool.conn.Send(pythia.Message{
		Message: pythia.DoneMsg,
		Id:      id,
		Status:  status,
		Output:  output,
	})
}

// Shut down the Pool component.
func (pool *Pool) Shutdown() {
	pool.conn.Close()
}

// vim:set sw=4 ts=4 noet:
