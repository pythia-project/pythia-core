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
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	job.interrupt = make(chan bool)
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
	job.output = strings.Replace(string(buffer[:read]), "\r\n", "\n", -1)
}

// Kill requests the sandbox to be killed. It is a convenience method to send
// a signal on the job.interrupt channel, but do not block if the job has
// already finished.
//
// Note: if kill() is called before the watch() goroutine is ready, the kill
// request will be silently ignored.
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

// Setup the parameters with the command line flags in args.
func (job *Job) Setup(args []string) {
	fs := flag.NewFlagSet(os.Args[0]+" execute", flag.ExitOnError)
	fs.StringVar(&job.UmlPath, "uml", "vm/uml", "path to the UML executable")
	fs.StringVar(&job.EnvDir, "envdir", "vm", "environments directory")
	fs.StringVar(&job.TasksDir, "tasksdir", "tasks", "tasks directory")
	if err := fs.Parse(args); err != nil {
		log.Fatal(err)
	}
	if len(fs.Args()) != 2 {
		log.Fatal("Usage: ", os.Args[0], " execute TASK INPUT")
	}
	taskfile, inputfile := fs.Arg(0), fs.Arg(1)
	taskcontent, err := ioutil.ReadFile(taskfile)
	if err != nil {
		log.Fatal(err)
	}
	if json.Unmarshal(taskcontent, &job.Task) != nil {
		log.Fatal(err)
	}
	inputcontent, err := ioutil.ReadFile(inputfile)
	if err != nil {
		log.Fatal(err)
	}
	job.Input = string(inputcontent)
}

// Execute the job when launched from the CLI. The result is shown on stdout.
func (job *Job) Run() {
	status, output := job.Execute()
	fmt.Println("Status:", status)
	fmt.Println("Output:", output)
}

// vim:set sw=4 ts=4 noet:
