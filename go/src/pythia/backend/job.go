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
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"pythia"
	"strings"
	"sync"
	"syscall"
	"time"
)

// A Job is the combination of a task and an input.
// Jobs are executed inside a sandbox.
//
// New jobs shall be created with the NewJob function.
//
// The Job type implements the pythia.Component interface so that a job can
// be launched from the CLI for debugging purposes.
type Job struct {
	Task  pythia.Task
	Input string

	// Path to the UML executable
	UmlPath string

	// Path to the directory containing the environments
	EnvDir string

	// Path to the directory containing the tasks
	TasksDir string

	// Process id of the job sandbox
	pid int

	// Job output
	output string

	// Send a signal to the interrupt channel to trigger job killing.
	interrupt chan bool

	// Has the job timed out, overflowed, or been aborted?
	timeout, overflow, abort bool

	// Error occurring in a goroutine
	err error

	// Barrier to wait for all goroutines associated to a job execution to end.
	wg sync.WaitGroup
}

// NewJob returns a new job, filled with default parameters. To execute the
// job, Task and Input have to be filled, and Execute() called.
func NewJob() *Job {
	job := new(Job)
	job.UmlPath = "vm/uml"
	job.EnvDir = "vm"
	job.TasksDir = "tasks"
	// The interrupt channel is buffered to avoid missing a kill request
	// arriving before the watch goroutine is ready.
	job.interrupt = make(chan bool, 1)
	return job
}

// Execute the job in a sandbox, wait for it to complete (or time out), and
// return the result.
func (job *Job) Execute() (status pythia.Status, output string) {
	// Write input to a temporary file. This is needed because UML has trouble
	// reading on the standard input. Hence, we feed the input as a block
	// device.
	inputfile, err := ioutil.TempFile("", "pythia-input-")
	if err != nil {
		return pythia.Error, fmt.Sprint(err)
	}
	defer os.Remove(inputfile.Name())
	defer inputfile.Close()
	if _, err := io.WriteString(inputfile, job.Input); err != nil {
		return pythia.Error, fmt.Sprint(err)
	}
	inputfile.Close()
	// Create and configure command.
	cmd := exec.Command(job.UmlPath,
		fmt.Sprintf("ubd0r=%s.sfs", path.Join(job.EnvDir, job.Task.Environment)),
		fmt.Sprintf("ubd1r=%s", path.Join(job.TasksDir, job.Task.TaskFS)),
		fmt.Sprintf("ubd2r=%s", inputfile.Name()),
		"con0=null,fd:1",
		"init=/init",
		"ro",
		"quiet",
		fmt.Sprintf("mem=%dm", job.Task.Limits.Memory),
		fmt.Sprintf("disksize=%d%%", job.Task.Limits.Disk))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = nil
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return pythia.Error, fmt.Sprint(err)
	}
	cmd.Stderr = cmd.Stdout
	// Run the VM
	if err := cmd.Start(); err != nil {
		return pythia.Error, fmt.Sprint(err)
	}
	job.pid = cmd.Process.Pid
	job.wg.Add(2)
	go job.watch()
	go job.gatherOutput(stdout)
	if err := cmd.Wait(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			// Ignore this error, cmd.ProcessState will be read below.
		default:
			job.err = err
		}
	}
	job.kill()
	job.wg.Wait()
	// Return result
	switch {
	case job.err != nil:
		return pythia.Error, fmt.Sprint(job.err)
	case job.abort:
		return pythia.Abort, job.output
	case job.overflow:
		return pythia.Overflow, job.output
	case job.timeout:
		return pythia.Timeout, job.output
	case !cmd.ProcessState.Success():
		return pythia.Crash, job.output
	default:
		return pythia.Success, job.output
	}
}

// Abort aborts the execution of the job.
func (job *Job) Abort() {
	job.abort = true
	job.kill()
}

// GatherOutput is a goroutine that buffers the output of the job.
// If the size of the output exceeds the limit set in the task, the job will be
// killed.
func (job *Job) gatherOutput(stdout io.Reader) {
	defer job.wg.Done()
	// Make buffer one byte larger than the limit to catch overflows.
	buffer := make([]byte, job.Task.Limits.Output+1)
	read := 0
	for {
		n, err := stdout.Read(buffer[read:])
		read += n
		if err != nil && err != io.EOF {
			job.err = err
			job.kill()
			break
		} else if read > job.Task.Limits.Output {
			job.overflow = true
			job.kill()
			break
		} else if err == io.EOF {
			break
		}
	}
	if read > job.Task.Limits.Output {
		read = job.Task.Limits.Output
	}
	// Trim NULL bytes that could occur at the end of the buffer
	if bytes.IndexByte(buffer, 0) != -1 {
		read = bytes.IndexByte(buffer, 0)
	}
	job.output = strings.Replace(string(buffer[:read]), "\r\n", "\n", -1)
}

// Kill requests the sandbox to be killed. It is a convenience method to send
// a signal on the job.interrupt channel, but do not block if the job has
// already finished.
func (job *Job) kill() {
	select {
	case job.interrupt <- true:
	default:
	}
}

// Watch is a goroutine responsible for killing the job when it times out or
// when a signal is received on the job.interrupt channel.
func (job *Job) watch() {
	defer job.wg.Done()
	select {
	case <-time.After(time.Duration(job.Task.Limits.Time) * time.Second):
		job.timeout = true
	case <-job.interrupt:
	}
	// Send the KILL signal to the whole UML process group. Do this even when
	// the job is already done to ensure no zombie processes are left.
	syscall.Kill(-job.pid, syscall.SIGKILL)
}

////////////////////////////////////////////////////////////////////////////////
// Component implementation for CLI debugging

func init() {
	pythia.Components["execute"] = pythia.ComponentInfo{
		Name:        "execute",
		Description: "Execute a single job (for debugging purposes)",
		New:         func() pythia.Component { return NewJob() },
	}
}

// Setup the parameters with the command line flags in args.
func (job *Job) Setup(fs *flag.FlagSet, args []string) error {
	taskfile := fs.String("task", "", "path to the task description (mandatory)")
	inputfile := fs.String("input", "", "path to the input file (mandatory)")
	fs.StringVar(&job.UmlPath, "uml", job.UmlPath, "path to the UML executable")
	fs.StringVar(&job.EnvDir, "envdir", job.EnvDir, "environments directory")
	fs.StringVar(&job.TasksDir, "tasksdir", job.TasksDir, "tasks directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(*taskfile) == 0 || len(*inputfile) == 0 {
		return errors.New("Missing task or input file")
	}
	taskcontent, err := ioutil.ReadFile(*taskfile)
	if err != nil {
		return err
	}
	if json.Unmarshal(taskcontent, &job.Task) != nil {
		return err
	}
	inputcontent, err := ioutil.ReadFile(*inputfile)
	if err != nil {
		return err
	}
	job.Input = string(inputcontent)
	return nil
}

// Execute the job when launched from the CLI. The result is shown on stdout.
func (job *Job) Run() {
	status, output := job.Execute()
	fmt.Println("Status:", status)
	fmt.Println("Output:", output)
}

// Abort the job if it is still running.
func (job *Job) Shutdown() {
	job.Abort()
}

// vim:set sw=4 ts=4 noet:
