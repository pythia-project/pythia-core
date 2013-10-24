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
// It checks that the execution time and output length are within the specified
// limits.
func runTask(t *testing.T, task *pythia.Task, input string) (status pythia.Status, output string) {
	job := newJob(task, input)
	wd := Watchdog(t, task.Limits.Time+1)
	status, output = job.Execute()
	wd.Stop()
	if len(output) > task.Limits.Output {
		t.Errorf("Job output is too large: max %d, got %d.", task.Limits.Output,
			len(output))
	}
	return
}

// RunTaskCheck behaves like RunTask, but additionally checks for expected
// status and output.
func runTaskCheck(t *testing.T, task *pythia.Task, input string,
	status pythia.Status, output string) {
	st, out := runTask(t, task, input)
	Expect(t, "status", status, st)
	Expect(t, "output", output, out)
}

// Shortcut for runTask(t, ReadTask(basename), ...)
func run(t *testing.T, basename string, input string, status pythia.Status,
	output string) {
	task, err := ReadTask(basename)
	if err != nil {
		t.Fatal(err)
	}
	runTaskCheck(t, task, input, status, output)
}

// Basic hello world task.
func TestJobHelloWorld(t *testing.T) {
	run(t, "hello-world", "", pythia.Success, "Hello world!\n")
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
	run(t, "timeout", "", pythia.Timeout, "Start\n")
}

// This task should overflow the output buffer.
func TestJobOverflow(t *testing.T) {
	task, err := ReadTask("overflow")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Trying limit 10")
	task.Limits.Output = 10
	runTaskCheck(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 6")
	task.Limits.Output = 6
	runTaskCheck(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 5")
	task.Limits.Output = 5
	runTaskCheck(t, task, "", pythia.Success, "abcde")
	t.Log("Trying limit 4")
	task.Limits.Output = 4
	runTaskCheck(t, task, "", pythia.Overflow, "abcd")
	t.Log("Trying limit 3")
	task.Limits.Output = 3
	runTaskCheck(t, task, "", pythia.Overflow, "abc")
}

// This task should overflow and be killed before the end.
func TestJobOverflowKill(t *testing.T) {
	wd := Watchdog(t, 2)
	run(t, "overflow-kill", "", pythia.Overflow, "abcde")
	wd.Stop()
}

// This task is a fork bomb. It should succeed, but not take the whole time.
func TestJobForkbomb(t *testing.T) {
	wd := Watchdog(t, 2)
	run(t, "forkbomb", "", pythia.Success, "Start\nDone\n")
	wd.Stop()
}

// Flooding the disk should not have any adverse effect.
func TestJobFlooddisk(t *testing.T) {
	run(t, "flooddisk", "", pythia.Success, "Start\nDone\n")
}

// vim:set sw=4 ts=4 noet:
