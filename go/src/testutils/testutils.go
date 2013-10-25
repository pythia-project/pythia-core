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

// Package testutils provides generic functions to simplify testing.
package testutils

import (
	"reflect"
	"testing"
	"time"
)

// Expect generates an error if expected and actual are not deeply equal.
func Expect(t *testing.T, name string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %s `%v`, got `%v`.", name, expected, actual)
	}
}

// Watchdog starts a timer that generates an error after seconds seconds, unless
// cancelled by its Stop method.
func Watchdog(t *testing.T, seconds int) *time.Timer {
	return time.AfterFunc(time.Duration(seconds)*time.Second, func() {
		t.Errorf("Time exceeded (%ds).", seconds)
	})
}

// vim:set sw=4 ts=4 noet:
