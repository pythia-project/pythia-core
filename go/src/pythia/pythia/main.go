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

// Main entry point
package main

import (
	"fmt"
	"os"
	"os/signal"
	"pythia"
	"pythia/backend"
	"syscall"
)

func usage() {
	fmt.Println("Usage:", os.Args[0], "component")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	name, args := os.Args[1], os.Args[2:]
	var component pythia.Component
	switch name {
	case "execute":
		component = backend.NewJob()
	case "pool":
		component = new(backend.Pool)
	default:
		fmt.Println("Unknown component", name)
		usage()
	}
	component.Setup(args)
	terminate, done := make(chan os.Signal, 1), make(chan bool, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-terminate:
			component.Shutdown()
		case <-done:
		}
	}()
	component.Run()
	done <- true
}

// vim:set sw=4 ts=4 noet:
