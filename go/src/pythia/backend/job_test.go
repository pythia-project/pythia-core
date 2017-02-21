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
	"testutils"
	"testutils/pytest"
	"time"
)

// NewTestJob creates a job, configured with the paths exported from make.
func newTestJob(task pythia.Task, input string) *Job {
	job := NewJob()
	job.Task = task
	job.Input = input
	job.UmlPath = pytest.UmlPath
	job.EnvDir = pytest.VmDir
	job.TasksDir = pytest.TasksDir
	return job
}

// RunTask executes task with input.
// It checks that the execution time and output length are within the specified
// limits.
func runTask(t *testing.T, task pythia.Task, input string) (status pythia.Status, output string) {
	job := newTestJob(task, input)
	wd := testutils.Watchdog(t, task.Limits.Time+1)
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
func runTaskCheck(t *testing.T, task pythia.Task, input string,
	status pythia.Status, output string) {
	st, out := runTask(t, task, input)
	testutils.Expect(t, "status", status, st)
	testutils.Expect(t, "output", output, out)
}

// Shortcut for runTask(t, pytest.ReadTask(t, basename), ...)
func run(t *testing.T, basename string, input string, status pythia.Status,
	output string) {
	runTaskCheck(t, pytest.ReadTask(t, basename), input, status, output)
}

// Basic hello world task.
func TestJobHelloWorld(t *testing.T) {
	run(t, "hello-world", "", pythia.Success, "Hello world!\n")
}

// Check that the goroutines are cleaned correctly.
func TestJobCleanup(t *testing.T) {
	testutils.CheckGoroutines(t, func() {
		run(t, "hello-world", "", pythia.Success, "Hello world!\n")
	})
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
	task := pytest.ReadTask(t, "overflow")
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
	wd := testutils.Watchdog(t, 2)
	run(t, "overflow-kill", "", pythia.Overflow, "abcde")
	wd.Stop()
}

// This task is a fork bomb. It should succeed, but not take the whole time.
func TestJobForkbomb(t *testing.T) {
	wd := testutils.Watchdog(t, 10)
	run(t, "forkbomb", "", pythia.Success, "Start\nDone\n")
	wd.Stop()
}

// Flooding the disk should not have any adverse effect.
func TestJobFlooddisk(t *testing.T) {
	run(t, "flooddisk", "", pythia.Success, "Start\nDone\n")
}

// Aborting a job shall be immediate.
func TestJobAbort(t *testing.T) {
	job := newTestJob(pytest.ReadTask(t, "timeout"), "")
	done := make(chan bool)
	go func() {
		wd := testutils.Watchdog(t, 2)
		status, output := job.Execute()
		wd.Stop()
		testutils.Expect(t, "status", pythia.Abort, status)
		testutils.Expect(t, "output", "Start\n", output)
		done <- true
	}()
	time.Sleep(1 * time.Second)
	job.Abort()
	<-done
}

// vim:set sw=4 ts=4 noet:
