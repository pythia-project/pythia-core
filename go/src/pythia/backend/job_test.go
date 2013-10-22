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
	"pythia"
	"strings"
	"testing"
	"time"
)

// NewJob creates a job, configured with the paths exported from make.
func newJob(task *pythia.Task, input string) Job {
	return Job{
		Task:     *task,
		Input:    input,
		UmlPath:  UmlPath,
		EnvDir:   VmDir,
		TasksDir: TasksDir,
	}
}

// RunTask executes task with input.
// It checks the result against the expected status and output (ignoring leading
// and trailing whitespace). The elapsed time running the task is returned.
func runTask(t *testing.T, task *pythia.Task, input string, status pythia.Status,
	output string) time.Duration {
	job := newJob(task, input)
	start := time.Now()
	st, out := job.Execute()
	elapsed := time.Since(start)
	if st != status {
		t.Errorf("Expected status `%s`, got `%s`.", status, st)
	}
	out = strings.TrimSpace(out)
	output = strings.TrimSpace(output)
	if out != output {
		t.Errorf("Expected output `%s`, got `%s`.", output, out)
	}
	if elapsed.Seconds() > float64(task.Limits.Time+1) {
		t.Errorf("Expected duration %ds, but task took %s.",
			task.Limits.Time, elapsed.String())
	}
	return elapsed
}

// Shortcut for runTask(t, ReadTask(basename), ...)
func run(t *testing.T, basename string, input string, status pythia.Status,
	output string) time.Duration {
	task, err := ReadTask(basename)
	if err != nil {
		t.Fatal(err)
	}
	return runTask(t, task, input, status, output)
}

// Basic hello world task.
func TestJobHelloWorld(t *testing.T) {
	run(t, "hello-world", "", pythia.Success, "Hello world!")
}

// Hello world task with input.
func TestJobHelloInput(t *testing.T) {
	run(t, "hello-input", "me\npythia\n",
		pythia.Success, "Hello me!\nHello pythia!\n")
}

// This task should time out after 5 seconds.
func TestJobTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	run(t, "timeout", "", pythia.Timeout, "Start")
}

// This task should overflow the output buffer.
func TestJobOverflow(t *testing.T) {
	task, err := ReadTask("overflow")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Trying limit 10")
	task.Limits.Output = 10
	runTask(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 6")
	task.Limits.Output = 6
	runTask(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 5")
	task.Limits.Output = 5
	runTask(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 4")
	task.Limits.Output = 4
	runTask(t, task, "", pythia.Overflow, "abcd")
	t.Log("Trying limit 3")
	task.Limits.Output = 3
	runTask(t, task, "", pythia.Overflow, "abc")
}

// This task should overflow and be killed before the end.
func TestJobOverflowKill(t *testing.T) {
	elapsed := run(t, "overflow-kill", "", pythia.Overflow, "abcde")
	if elapsed.Seconds() > 2 {
		t.Errorf("Task took too long: %s.", elapsed.String())
	}
}

// This task is a fork bomb. It should succeed, but not take the whole time.
func TestJobForkbomb(t *testing.T) {
	elapsed := run(t, "forkbomb", "", pythia.Success, "Start\nDone")
	if elapsed.Seconds() > 2 {
		t.Errorf("Task took too long: %s.", elapsed.String())
	}
}

// Flooding the disk should not have any adverse effect.
func TestJobFlooddisk(t *testing.T) {
	run(t, "flooddisk", "", pythia.Success, "Start\nDone")
}

// vim:set sw=4 ts=4 noet:
