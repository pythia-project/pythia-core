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
	"flag"
)

// A Component is meant to be run in its own process, and maybe on a separate
// machine.
type Component interface {
	// Parse the command line arguments given by args and configure the
	// component.
	Setup(fs *flag.FlagSet, args []string) error

	// Execute the component.
	Run()

	// Shut down the component.
	Shutdown()
}

// A ComponentInfo is a type of component. Each type registers in the global
// Components map.
type ComponentInfo struct {
	// CLI name of the component (must be unique).
	// This is also the key of the Components map.
	Name string

	// Short one-line description that can be shown in the CLI.
	Description string

	// Function to create a new component of this type.
	New func() Component
}

// The global Components map contains all registered component types.
var Components = make(map[string]ComponentInfo)

// vim:set sw=4 ts=4 noet:
