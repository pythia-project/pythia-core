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
	// The address on which the queue listens.
	QueueAddr, _ = ParseAddr("127.0.0.1:9000")

	// The interval at which to send keep-alive messages on idle connections.
	// This setting shall be set before any Conn has been created, and shall not
	// be altered afterwards.
	KeepAliveInterval = 30 * time.Second

	// Initial interval between dial tries. After each failed attempt, the
	// interval is doubled, up to MaxRetryInterval.
	InitialRetryInterval = 32 * time.Millisecond

	// Maximum time interval between dial tries.
	MaxRetryInterval = 5 * time.Minute
)

// vim:set sw=4 ts=4 noet:
