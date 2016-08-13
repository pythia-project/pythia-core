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
// Warning: This function should not used except for testing purpose
func Run(t *testing.T, basename string, input string, status pythia.Status,
	output string) {
	runTaskCheck(t, pytest.ReadTask(t, basename), input, status, output)
}
