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

package backend

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"pythia"
	"strings"
	"testing"
)

// Environment variables exported from make
var (
	topDir   = os.Getenv("TOP_DIR")
	outDir   = path.Join(topDir, os.Getenv("OUT_DIR"))
	umlPath  = path.Join(topDir, os.Getenv("UML"))
	vmDir    = path.Join(topDir, os.Getenv("VM_OUT_DIR"))
	tasksDir = path.Join(topDir, os.Getenv("TASKS_OUT_DIR"))
)

// ReadTask reads a task description from a file.
func readTask(basename string) (*pythia.Task, error) {
	content, err := ioutil.ReadFile(path.Join(tasksDir, basename+".task"))
	if err != nil {
		return nil, err
	}
	task := new(pythia.Task)
	if json.Unmarshal(content, task) != nil {
		return nil, err
	}
	return task, nil
}

// NewJob creates a job, configured with the paths exported from make.
func newJob(task *pythia.Task, input string) Job {
	return Job{
		Task:     *task,
		Input:    input,
		UmlPath:  umlPath,
		EnvDir:   vmDir,
		TasksDir: tasksDir,
	}
}

// RunTask reads the task description basename and executes it with input.
// It checks the result against the expected status and output (ignoring leading
// and trailing whitespace).
func runTask(t *testing.T, basename string, input string, status pythia.Status, output string) {
	task, err := readTask(basename)
	if err != nil {
		t.Fatal(err)
	}
	job := newJob(task, input)
	st, out := job.Execute()
	if st != status {
		t.Errorf("Expected status `%s`, got `%s`.", status, st)
	}
	out = strings.TrimSpace(out)
	output = strings.TrimSpace(output)
	if out != output {
		t.Errorf("Expected output `%s`, got `%s`.", output, out)
	}
}

// Basic hello world task.
func TestHelloWorld(t *testing.T) {
	runTask(t, "hello-world", "", pythia.Success, "Hello world!")
}

// Hello world task with input.
func TestHelloInput(t *testing.T) {
	runTask(t, "hello-input", "me\npythia\n",
		pythia.Success, "Hello me!\nHello pythia!\n")
}

// vim:set sw=4 ts=4 noet:
