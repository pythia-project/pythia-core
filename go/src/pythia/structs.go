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
	"encoding/json"
)

// Status of a task execution.
type Status string

const (
	Success  Status = "success"  // finished, output = stdout
	Timeout  Status = "timeout"  // timed out, output = stdout so far
	Overflow Status = "overflow" // stdout too big, output = capped stdout
	Abort    Status = "abort"    // aborted by abort message, no output
	Crash    Status = "crash"    // sandbox crashed, output = stdout
	Error    Status = "error"    // (maybe temporary) error, output = error message
	Fatal    Status = "fatal"    // unrecoverable error (e.g. misformatted task), output = error message
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

func (task Task) String() string {
	s, err := json.Marshal(task)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(s)
}

// Message type.
// Components may have internal message types starting with a hyphen.
type MsgType string

const (
	// Internal message for keeping a connection alive.
	KeepAliveMsg MsgType = "keep-alive"

	// Register a sandbox pool.
	// Pool->Queue
	RegisterPoolMsg MsgType = "register-pool"

	// Request execution of a task.
	// Frontend->Queue, Queue->Pool
	LaunchMsg MsgType = "launch"

	// Job done.
	// Pool->Queue, Queue->Frontend.
	DoneMsg MsgType = "done"

	// Abort job. The receiving end shall send a done message with status abort
	// (or another status if the job has ended meanwhile).
	// Frontend->Queue, Queue->Pool
	AbortMsg MsgType = "abort"
)

// A Message is the basic entity that is sent between components. Messages are
// serialized to JSON.
type Message struct {
	// The message.
	Message MsgType `json:"message"`

	// The capacity of the pool. Only for message register-pool.
	Capacity int `json:"capacity,omitempty"`

	// The task identifier. Only for messages launch, done and abort.
	Id string `json:"id,omitempty"`

	// The task to launch. Only for message launch.
	Task *Task `json:"task,omitempty"`

	// The input to feed to the task. Only for message launch.
	Input string `json:"input,omitempty"`

	// The result status of the execution. Only for message done.
	Status Status `json:"status,omitempty"`

	// The result output of the execution. Only for message done.
	Output string `json:"output,omitempty"`
}

func (msg Message) String() string {
	s, err := json.Marshal(msg)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(s)
}

// vim:set sw=4 ts=4 noet:
