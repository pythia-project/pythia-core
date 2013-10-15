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

// A Component is meant to be run in its own process, and maybe on a separate
// machine.
type Component interface {
	// Parse the command line arguments given by args and configure the
	// component.
	Setup(args []string)

	// Execute the component.
	Run()
}

// vim:set sw=4 ts=4 noet:
