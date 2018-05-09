// Copyright 2016 The Pythia Authors.
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

package frontend

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pythia"
	"strconv"
	"syscall"
)

func init() {
	pythia.Components["server"] = pythia.ComponentInfo{
		Name:        "server",
		Description: "Front-end component allowing execution of pythia tasks",
		New:         func() pythia.Component { return NewServer() },
	}
}

// A taskRequest is an internal structure to hold the request sent by the client
// It contains the information about the task to be executed.
type taskRequest struct {
	// The task identifier.
	Tid string

	// The input to be used for the task execution.
	Response string
}

// A Server is a component that allows client to execute tasks.
//
// New servers shall be created by the NewServer function.
//
// A Server connects to the Queue to execute the task, waits for its complete
// execution and sends back the result to the client.
type Server struct {
	// The port number on which this server is listening.
	Port int
}

// NewServer returns a new server with default parameters.
func NewServer() *Server {
	server := new(Server)
	server.Port = 8080
	return server
}

// Setup configures the server with the command line flags in args.
func (server *Server) Setup(fs *flag.FlagSet, args []string) error {
	fs.IntVar(&server.Port, "port", server.Port, "server port")
	return fs.Parse(args)
}

// Run the Server component.
func (server *Server) Run() {
	// Catch ctrl+c and SIGTERM
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		signalType := <-ch
		log.Println("Received signal", signalType)
		signal.Stop(ch)
		os.Exit(0)
	}()
	// Start the web server
	router := NewRouter()
	httpServ := &http.Server{
		Addr:    ":" + strconv.Itoa(server.Port),
		Handler: MiddleWare(router),
	}

	log.Println("Server listening on", server.Port)
	if err := httpServ.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Shut down the Server component.
func (server *Server) Shutdown() {
}

// vim:set sw=4 ts=4 noet:
