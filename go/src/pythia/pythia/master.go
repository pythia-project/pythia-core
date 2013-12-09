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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pythia"
	"strings"
	"sync"
	"syscall"
	"time"
)

// A message line to log
type logMessage struct {
	Message   string
	Component pythia.ComponentInfo
	Pid       int
}

// The component info structure for the master pseudo-component.
// Note that this component is not added to the global component registry.
var masterInfo = pythia.ComponentInfo{
	Name:        "master",
	Description: "Master process",
	New:         func() pythia.Component { return NewMaster() },
}

// Master is a special (unregistered) component that manages the lifecycle of
// other components. The components are defined in the configuration file.
// Each component runs in its own process and is restarted if it fails.
type Master struct {
	// The log channel where messages shall be submitted
	log chan<- logMessage

	// Channel to signal a shutdown
	quit chan bool

	// Whether we are in the process of shutting down
	quitting bool

	// Flags for exit code
	errFatal, errUsage bool
}

// NewMaster returns a new Master component.
func NewMaster() *Master {
	master := new(Master)
	master.quit = make(chan bool, 1)
	return master
}

// Setup is implemented to satisfy the pythia.Component interface, but has no
// function since the components are configured in the config file.
func (master *Master) Setup(fs *flag.FlagSet, args []string) error {
	return nil
}

// Run runs the Master component.
func (master *Master) Run() {
	// Set up log facility
	log := make(chan logMessage)
	master.log = log
	go master.logger(log)
	// Launch all components
	var wg sync.WaitGroup
	for _, config := range ComponentConfigs {
		wg.Add(1)
		go func(config ComponentConfig) {
			comp := masterComponent{Master: master, Config: config}
			comp.Run()
			wg.Done()
		}(config)
	}
	// Wait for all components to finish.
	wg.Wait()
	close(log)
	if master.errFatal {
		os.Exit(1)
	} else if master.errUsage {
		os.Exit(2)
	}
}

// Shutdown requests a shutdown of the whole system.
func (master *Master) Shutdown() {
	master.quitting = true
	select {
	case master.quit <- true:
	default:
	}
}

// The logger goroutine writes log messages to the standard output.
// Having a centralized place for this ensures lines don't get mixed up.
func (master *Master) logger(log chan logMessage) {
	for m := range log {
		fmt.Printf("%s %s [%d]: %s\n", time.Now().Format("2006-01-02 15:04:05"),
			m.Component.Name, m.Pid, m.Message)
	}
}

// Log message msg originating from the master process.
func (master *Master) Log(msg string) {
	master.log <- logMessage{
		Message:   msg,
		Component: masterInfo,
		Pid:       os.Getpid(),
	}
}

// Error logs an error, sets the fatal error flag and requests a shutdown.
func (master *Master) Error(err error) {
	master.Log(err.Error())
	master.errFatal = true
	master.Shutdown()
}

// A masterComponent centralizes the operations on one component running in its
// own process.
type masterComponent struct {
	Master *Master
	Config ComponentConfig
	pid    int
	done   chan bool
	wg     sync.WaitGroup
}

// Run runs the component in its own process, restarting it when needed.
func (comp *masterComponent) Run() {
	comp.done = make(chan bool, 1)
	// Construct command line
	args := append(os.Args[1:], comp.Config.Info.Name)
	for key, value := range comp.Config.Options {
		args = append(args, "-"+key, value)
	}
	for !comp.Master.quitting {
		comp.Master.Log("Launching " + strings.Join(args[len(os.Args)-1:], " "))
		cmd := exec.Command(os.Args[0], args...)
		cmd.Stdin = nil
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			comp.Master.Error(err)
			return
		}
		cmd.Stderr = cmd.Stdout
		// Execute component
		if err := cmd.Start(); err != nil {
			comp.Master.Error(err)
			return
		}
		comp.pid = cmd.Process.Pid
		comp.wg = sync.WaitGroup{}
		comp.wg.Add(2)
		go comp.gatherOutput(stdout)
		go comp.watch()
		// Wait for finish
		if err := cmd.Wait(); err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				// Handled below
			default:
				comp.Master.Error(err)
			}
		}
		comp.Log("Exited with " + cmd.ProcessState.String())
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if status.Exited() && status.ExitStatus() == 2 {
			comp.Master.errUsage = true
			comp.Master.Shutdown()
		}
		comp.done <- true
		comp.wg.Wait()
	}
}

// Log message msg originating from this component.
func (comp *masterComponent) Log(msg string) {
	comp.Master.log <- logMessage{
		Message:   msg,
		Component: comp.Config.Info,
		Pid:       comp.pid,
	}
}

// The gatherOutput goroutine gathers the output of the process and sends it
// line-by-line to the master logger.
func (comp *masterComponent) gatherOutput(stdout io.Reader) {
	defer comp.wg.Done()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		comp.Log(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		comp.Master.Error(err)
	}
}

// The watch goroutine terminates the component when a shutdown is requested.
func (comp *masterComponent) watch() {
	defer comp.wg.Done()
	select {
	case <-comp.done:
		return
	case <-comp.Master.quit:
		comp.Master.quit <- true
		comp.Log("Terminating")
		syscall.Kill(comp.pid, syscall.SIGTERM)
	}
	select {
	case <-comp.done:
		return
	case <-time.After(5 * time.Second):
		comp.Log("Killing")
		syscall.Kill(comp.pid, syscall.SIGKILL)
	}
}

// vim:set sw=4 ts=4 noet:
