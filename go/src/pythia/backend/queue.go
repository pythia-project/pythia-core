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
	"container/list"
	"flag"
	"fmt"
	"log"
	"pythia"
	"strings"
	"sync"
	"time"
)

func init() {
	pythia.Components["queue"] = pythia.ComponentInfo{
		Name:        "queue",
		Description: "Central queue back-end component",
		New:         func() pythia.Component { return NewQueue() },
	}
}

// A queueClient is an internal structure keeping information about active
// connections. An active connection may be a pool or a client (or even both).
// We consider each connection to be a pool (maybe with Capacity 0).
type queueClient struct {
	// Unique identifier of the connection
	Id int

	// The response channel.
	Response chan<- pythia.Message

	// The number of parallel jobs this pool can handle.
	Capacity int

	// Jobs currently running in this pool, mapped by job id.
	Running map[string]*queueJob

	// Jobs submitted (and not yet done) by this client, mapped by job id.
	Submitted map[string]*queueJob
}

// A queueJob is an internal structure keeping information about a job during
// its whole lifetime in the queue.
//
// Invariant: WaitingElement != nil || Pool != nil
type queueJob struct {
	// The job identifier. Must be the same as Msg.Id.
	Id string

	// The launch message.
	Msg pythia.Message

	// The client having submitted this job.
	Origin *queueClient

	// Element of the queue.Waiting list pointing to this job, or nil if the job
	// is currently running.
	WaitingElement *list.Element

	// Pool in which this job is currently running, or nil if the job is waiting
	// to be scheduled.
	Pool *queueClient
}

// A queueMessage is an internal message from a queue connection handler to the
// queue main goroutine. It contains the (possibly altered) message and the
// originating client.
type queueMessage struct {
	Msg    pythia.Message
	Client *queueClient
}

// Internal messages
const (
	// A client has connected
	connectMsg pythia.MsgType = "-connect"

	// A connection has closed
	closedMsg pythia.MsgType = "-closed"

	// Shutdown has been requested
	quitMsg pythia.MsgType = "-quit"
)

// The Queue is the central component of Pythia.
// It receives jobs (tasks with inputs) from front-ends and dispatches them
// to the sandboxes.
//
// The Queue is the only component listening for connections. All other
// components connect to it.
type Queue struct {
	// The maximum number of jobs that can wait to be executed.
	Capacity int `json:"capacity"`

	// Channel to send messages to the main goroutine
	master chan<- queueMessage

	// Channel to request shutdown
	quit chan bool

	// WaitGroup for all goroutines
	wg sync.WaitGroup

	// Active connections
	Clients map[int]*queueClient `json:"clients"`

	// Jobs to be processed/currently processing
	Jobs map[string]*queueJob `json:"jobs"`

	// List of jobs (*queueJob) waiting to be assigned.
	Waiting *list.List `json:"waiting"`

	// Get the Queue creation datetime
	CreationDate Time `json:"creation_date"`
}

// NewQueue returns a new queue with default parameters.
func NewQueue() *Queue {
	queue := new(Queue)
	queue.Capacity = 500
	queue.quit = make(chan bool, 1)
	queue.CreationDate = time.Now()
	return queue
}

// Setup configures the queue with the command line flags in args.
func (queue *Queue) Setup(fs *flag.FlagSet, args []string) error {
	fs.IntVar(&queue.Capacity, "capacity", queue.Capacity, "queue capacity")
	return fs.Parse(args)
}

// Run runs the Queue component.
func (queue *Queue) Run() {
	l, err := pythia.Listen(pythia.QueueAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening to", pythia.QueueAddr)
	closing := false
	master := make(chan queueMessage)
	queue.master = master
	go func() {
		<-queue.quit
		closing = true
		l.Close()
	}()
	queue.wg.Add(1)
	go queue.main(master)
	nextid := 0
	for {
		conn, err := l.Accept()
		if closing {
			break
		} else if err != nil {
			log.Print(err)
			continue
		}
		response := make(chan pythia.Message)
		client := &queueClient{
			Id:        nextid,
			Response:  response,
			Running:   make(map[string]*queueJob),
			Submitted: make(map[string]*queueJob),
		}
		master <- queueMessage{pythia.Message{Message: connectMsg}, client}
		queue.wg.Add(1)
		go queue.handle(conn, client, response)
		nextid++
	}
	master <- queueMessage{pythia.Message{Message: quitMsg}, nil}
	queue.wg.Wait()
}

// Shutdown terminates the Queue component.
func (queue *Queue) Shutdown() {
	select {
	case queue.quit <- true:
	default:
	}
}

// Main goroutine responsible for scheduling the jobs.
func (queue *Queue) main(master <-chan queueMessage) {
	defer queue.wg.Done()
	queue.Clients = make(map[int]*queueClient)
	queue.Jobs = make(map[string]*queueJob)
	queue.Waiting = list.New()
	for qm := range master {
		switch qm.Msg.Message {
		case connectMsg:
			log.Print("Client ", qm.Client.Id, ": connected.")
			queue.Clients[qm.Client.Id] = qm.Client
		case pythia.RegisterPoolMsg:
			log.Print("Client ", qm.Client.Id, ": pool capacity ",
				qm.Msg.Capacity)
			qm.Client.Capacity = qm.Msg.Capacity
		case pythia.LaunchMsg:
			id := qm.Msg.Id
			if _, ok := queue.Jobs[id]; ok {
				log.Print("Job ", id, ": already launched, rejecting.")
				qm.Client.Response <- pythia.Message{
					Message: pythia.DoneMsg,
					Id:      id,
					Status:  pythia.Fatal,
					Output:  "Job already launched",
				}
			} else if queue.Waiting.Len() >= queue.Capacity {
				log.Print("Job ", id, ": queue full, rejecting.")
				qm.Client.Response <- pythia.Message{
					Message: pythia.DoneMsg,
					Id:      id,
					Status:  pythia.Error,
					Output:  "Queue full",
				}
			} else {
				job := &queueJob{
					Id:     id,
					Msg:    qm.Msg,
					Origin: qm.Client,
				}
				qm.Client.Submitted[id] = job
				queue.Jobs[id] = job
				job.WaitingElement = queue.Waiting.PushBack(job)
				log.Print("Job ", id, ": queued.")
			}
		case pythia.DoneMsg:
			id := qm.Msg.Id
			log.Print("Job ", id, ": done.")
			job := queue.Jobs[id]
			if job == nil {
				log.Println("Ignoring message for unknown job", qm.Msg)
				break
			}
			pool := job.Pool
			if pool == nil || pool != qm.Client {
				log.Println("Ignoring message from wrong source", qm.Msg)
				break
			}
			delete(queue.Jobs, id)
			delete(pool.Running, id)
			if job.Origin != nil {
				// job.Origin is nil if the submitting client has disconnected
				// before receiving the result.
				delete(job.Origin.Submitted, id)
				job.Origin.Response <- qm.Msg
			}
		case closedMsg:
			log.Print("Client ", qm.Client.Id, ": disconnected.")
			close(qm.Client.Response)
			delete(queue.Clients, qm.Client.Id)
			for _, job := range qm.Client.Running {
				if job.Origin == nil {
					// Submitter disconnected, we can discard the job.
					delete(queue.Jobs, job.Id)
				} else {
					// Otherwise, reschedule it.
					job.Pool = nil
					job.WaitingElement = queue.Waiting.PushFront(job)
				}
			}
			for _, job := range qm.Client.Submitted {
				if job.WaitingElement != nil {
					// Job is in waiting queue, discard it.
					queue.Waiting.Remove(job.WaitingElement)
					delete(queue.Jobs, job.Id)
				} else if job.Pool != nil {
					// Job is running, abort it.
					job.Origin = nil
					job.Pool.Response <- pythia.Message{
						Message: pythia.AbortMsg,
						Id:      job.Id,
					}
					// Keep job in queue.Jobs to handle abort result
				}
			}
		case quitMsg:
			log.Println("Quitting.")
			goto quit
		default:
			log.Fatal("Invalid internal message", qm.Msg)
		}

		// Schedule jobs
		queue.schedule()
	}

quit:
	if len(queue.Clients) == 0 {
		return
	}
	for _, client := range queue.Clients {
		close(client.Response)
	}
	// Wait for all Clients to quit. We flush messages from the master channel
	// to ensure no connection handler is in a deadlock.
	for qm := range master {
		switch qm.Msg.Message {
		case closedMsg:
			delete(queue.Clients, qm.Client.Id)
			if len(queue.Clients) == 0 {
				return
			}
		default:
			// Swallow all other messages
		}
	}
}

// Schedule assigns waiting jobs to free sandboxes.
// This function shall be called from the main goroutine, as it manipulates
// the queue data structures.
func (queue *Queue) schedule() {
	if queue.Waiting.Len() == 0 {
		return
	}
	for _, client := range queue.Clients {
		for len(client.Running) < client.Capacity {
			job := queue.Waiting.Remove(queue.Waiting.Front()).(*queueJob)
			job.WaitingElement = nil
			job.Pool = client
			client.Running[job.Id] = job
			client.Response <- job.Msg
			if queue.Waiting.Len() == 0 {
				return
			}
		}
	}
}

// Handle the connection with another component (front-end or pool).
// All job ids are prepended by the client id to ensure unique ids.
func (queue *Queue) handle(conn *pythia.Conn, client *queueClient, response chan pythia.Message) {
	defer queue.wg.Done()
	defer conn.Close()
	queue.wg.Add(1)
	go func() {
		// Receiver goroutine: reads messages from the client and send them to
		// the main goroutine.
		defer queue.wg.Done()
		defer func() { queue.master <- queueMessage{pythia.Message{Message: closedMsg}, client} }()
		for msg := range conn.Receive() {
			switch msg.Message {
			case pythia.RegisterPoolMsg:
				if msg.Capacity < 1 {
					log.Println("Invalid pool capacity", msg.Capacity)
				} else {
					queue.master <- queueMessage{msg, client}
				}
			case pythia.LaunchMsg:
				msg.Id = fmt.Sprintf("%d:%s", client.Id, msg.Id)
				queue.master <- queueMessage{msg, client}
			case pythia.DoneMsg:
				queue.master <- queueMessage{msg, client}
			default:
				log.Println("Ignoring message", msg)
			}
		}
	}()
	// Handle responses from the main goroutine and send messages to the client.
	for msg := range response {
		switch msg.Message {
		case pythia.LaunchMsg:
			conn.Send(msg)
		case pythia.DoneMsg:
			msg.Id = msg.Id[strings.Index(msg.Id, ":")+1:]
			conn.Send(msg)
		default:
			log.Fatal("Invalid internal message", msg)
		}
	}
}

// vim:set ts=4 sw=4 noet:
