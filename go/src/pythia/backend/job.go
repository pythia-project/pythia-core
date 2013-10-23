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

	// Has a timeout or overflow occurred?
	timeout, overflow bool

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
	kill, done := make(chan bool, 1), make(chan bool, 1)
	if err := cmd.Start(); err != nil {
		return pythia.Error, fmt.Sprint(err)
	}
	job.pid = cmd.Process.Pid
	job.wg.Add(2)
	go job.gatherOutput(stdout, kill)
	go job.watch(kill, done)
	if err := cmd.Wait(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			// Ignore this error, cmd.ProcessState will be read below.
		default:
			job.err = err
		}
	}
	done <- true
	job.wg.Wait()
	// Return result
	switch {
	case job.err != nil:
		return pythia.Error, fmt.Sprint(job.err)
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

// GatherOutput is a goroutine that buffers the output of the job.
// If the size of the output exceeds the limit set in the task, the job will be
// killed by signaling on the kill channel.
func (job *Job) gatherOutput(stdout io.Reader, kill chan<- bool) {
	defer job.wg.Done()
	// Make buffer one byte larger than the limit to catch overflows.
	buffer := make([]byte, job.Task.Limits.Output+1)
	read := 0
	for {
		n, err := stdout.Read(buffer[read:])
		read += n
		if err != nil && err != io.EOF {
			job.err = err
			kill <- true
			break
		} else if read > job.Task.Limits.Output {
			job.overflow = true
			kill <- true
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

// Kill kills the sandbox. We send the KILL signal to the whole process group.
func (job *Job) kill() {
	syscall.Kill(-job.pid, syscall.SIGKILL)
}

// Watch is a goroutine responsible for killing the job when it times out or
// when a signal is received on the kill channel. A signal on the done channel
// means the job has exited and nothing has to be done anymore.
func (job *Job) watch(kill, done <-chan bool) {
	defer job.wg.Done()
	select {
	case <-time.After(time.Duration(job.Task.Limits.Time) * time.Second):
		job.timeout = true
		job.kill()
	case <-kill:
		job.kill()
	case <-done:
	}
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
