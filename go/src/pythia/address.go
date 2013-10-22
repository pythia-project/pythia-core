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
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// ParseAddr parses an address description into an address.
// If the description starts with "unix:", the result will be a Unix domain
// socket address with the path being the rest of the description string.
// Otherwise, the result will be a TCP address represented by the whole
// description string.
func ParseAddr(description string) (net.Addr, error) {
	if strings.HasPrefix(description, "unix:") {
		return net.ResolveUnixAddr("unix", description[len("unix:"):])
	} else {
		return net.ResolveTCPAddr("tcp", description)
	}
}

// LocalAddr returns a random local unix address located in the system's
// temporary directory.
func LocalAddr() (net.Addr, error) {
	f, err := ioutil.TempFile("", "pythia.sock-")
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(f.Name()); err != nil {
		return nil, err
	}
	addr, err := net.ResolveUnixAddr("unix", f.Name())
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// vim:set sw=4 ts=4 noet:
