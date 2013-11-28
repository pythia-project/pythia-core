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
	"time"
)

// Global configuration
var (
	// The interval at which to send keep-alive messages on idle connections.
	// This setting shall be set before any Conn has been created, and shall not
	// be altered afterwards.
	KeepAliveInterval = 30 * time.Second
)

// vim:set sw=4 ts=4 noet:
