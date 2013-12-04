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
	"os"
	"pythia"
	"sync"
)

func init() {
	pythia.Components["pool"] = pythia.ComponentInfo{
		Name:        "pool",
		Description: "Back-end component managing a pool of sandboxes",
		New:         func() pythia.Component { return NewPool() },
	}
}

// A Pool is a component that launches sandboxes on the local machine.
// Each Pool has a limit on the number of sandboxes that can run concurrently.
//
// New pools shall be created by the NewPool function.
//
// A Pool connects to the Queue, advertises its limits, and waits for jobs to
// execute.
type Pool struct {
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

	// Channel to request shutdown
	quit chan bool

	// Channel to abort all remaining jobs
	abort chan bool
}

// NewPool returns a new pool with default parameters.
func NewPool() *Pool {
	pool := new(Pool)
	pool.Capacity = 1
	pool.UmlPath = "vm/uml"
	pool.EnvDir = "vm"
	pool.TasksDir = "tasks"
	pool.quit = make(chan bool, 1)
	return pool
}

// Setup the parameters with the command line flags in args.
func (pool *Pool) Setup(args []string) {
	fs := flag.NewFlagSet(os.Args[0]+" pool", flag.ExitOnError)
	fs.IntVar(&pool.Capacity, "capacity", pool.Capacity, "max parallel sandboxes")
	fs.StringVar(&pool.UmlPath, "uml", pool.UmlPath, "path to the UML executable")
	fs.StringVar(&pool.EnvDir, "envdir", pool.EnvDir, "environments directory")
	fs.StringVar(&pool.TasksDir, "tasksdir", pool.TasksDir, "tasks directory")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
}

// Run the Pool component.
func (pool *Pool) Run() {
	conn := pythia.DialRetry(pythia.QueueAddr)
	defer conn.Close()
	pool.conn = conn
	// Tokens is a buffered channel to enforce the capacity. Values do not
	// matter.
	tokens := make(chan bool, pool.Capacity)
	for i := 0; i < pool.Capacity; i++ {
		tokens <- true
	}
	pool.abort = make(chan bool, 1)
	var wg sync.WaitGroup
	conn.Send(pythia.Message{
		Message:  pythia.RegisterPoolMsg,
		Capacity: pool.Capacity,
	})
mainloop:
	for {
		select {
		case msg, ok := <-conn.Receive():
			if !ok {
				break mainloop
			}
			switch msg.Message {
			case pythia.LaunchMsg:
				select {
				case <-tokens:
					wg.Add(1)
					go func(msg pythia.Message) {
						pool.doJob(msg.Id, msg.Task, msg.Input)
						tokens <- true
						wg.Done()
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
		case <-pool.quit:
			break mainloop
		}
	}
	conn.Close()
	pool.abort <- true
	wg.Wait()
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
	done := make(chan bool)
	go func() {
		status, output := job.Execute()
		pool.conn.Send(pythia.Message{
			Message: pythia.DoneMsg,
			Id:      id,
			Status:  status,
			Output:  output,
		})
		done <- true
	}()
	select {
	case <-pool.abort:
		pool.abort <- true
		job.Abort()
		<-done
	case <-done:
	}
}

// Shut down the Pool component.
func (pool *Pool) Shutdown() {
	select {
	case pool.quit <- true:
	default:
	}
}

// vim:set sw=4 ts=4 noet:
