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

// Package pytest provides functions to simplify testing in Pythia.
package pytest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"pythia"
	"testing"
)

// Environment variables exported from make
var (
	TopDir   = os.Getenv("TOP_DIR")
	OutDir   = path.Join(TopDir, os.Getenv("OUT_DIR"))
	UmlPath  = path.Join(TopDir, os.Getenv("UML"))
	VmDir    = path.Join(TopDir, os.Getenv("VM_OUT_DIR"))
	TasksDir = path.Join(TopDir, os.Getenv("TASKS_OUT_DIR"))
)

// ReadTask reads a task description from a file.
//
// If an error occurs while reading the task description, the t.Fatal() will
// be called. Hence, ReadTask must be called from the main test goroutine.
func ReadTask(t *testing.T, basename string) (task pythia.Task) {
	content, err := ioutil.ReadFile(path.Join(TasksDir, basename+".task"))
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal(content, &task) != nil {
		t.Fatal(err)
	}
	return
}

// vim:set sw=4 ts=4 noet:
