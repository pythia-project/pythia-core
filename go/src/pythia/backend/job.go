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
	"syscall"
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
}

// Execute the job in a sandbox, wait for it to complete (or time out), and
// return the result.
func (job *Job) Execute() (status pythia.Status, output string) {
	// BUG(vianney): Not all limits are currently enforced during job execution.
	inputfile, err := ioutil.TempFile("", "pythia")
	if err != nil {
		return pythia.Crash, fmt.Sprint(err)
	}
	defer os.Remove(inputfile.Name())
	defer inputfile.Close()
	if _, err := io.WriteString(inputfile, job.Input); err != nil {
		return pythia.Crash, fmt.Sprint(err)
	}
	inputfile.Close()
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
	out, err := cmd.Output()
	if err != nil {
		return pythia.Crash, fmt.Sprint(err)
	}
	return pythia.Success, strings.Replace(string(out), "\r\n", "\n", -1)
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
