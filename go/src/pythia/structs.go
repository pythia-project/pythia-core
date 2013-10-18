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

// Status of a task execution.
type Status string

const (
	Success  Status = "success"  // finished, output = stdout
	Timeout         = "timeout"  // timed out, output = stdout so far
	Overflow        = "overflow" // stdout too big, output = capped stdout
	Abort           = "abort"    // aborted by abort message, no output
	Crash           = "crash"    // sandbox crashed, output = stdout + exit code
	Error           = "error"    // unrecoverable error (e.g. misformatted task), output = error message
)

// Task is the description of a task to be run in a sandbox.
type Task struct {
	// Environment is the name of the root filesystem.
	Environment string `json:"environment"`

	// TaskFS is the relative path to the task filesystem.
	TaskFS string `json:"taskfs"`

	// Execution limits to be enforced in the sandbox.
	Limits struct {
		// Maximum execution time in seconds.
		Time int `json:"time"`

		// Total amount of main memory (in megabytes) allocated
		// to the sandbox VM.
		Memory int `json:"memory"`

		// Fraction (in percents) of main memory that can be used as disk space
		// in a tmpfs. Note that only used disk space is allocated.
		Disk int `json:"disk"`

		// Maximum size of the output (in bytes).
		Output int `json:"output"`
	} `json:"limits"`
}

// vim:set sw=4 ts=4 noet:
