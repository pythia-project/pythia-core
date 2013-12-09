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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"pythia"
	_ "pythia/backend"
)

// Config is the structure of the configuration file.
type Config struct {
	// Global options
	Global map[string]string

	// List of components to run in master mode
	Components []map[string]string
}

// A ComponentConfig is a component that is configured in the configuration
// file.
type ComponentConfig struct {
	// The registered component information
	Info pythia.ComponentInfo

	// The options
	Options map[string]string
}

var (
	// The list of components configured in the configuration file.
	// This list is populated by ParseConfig()
	ComponentConfigs []ComponentConfig

	// The configuration filename
	ConfigFile string

	// The global flag set
	gfs *flag.FlagSet

	// Proxy variable for pythia.QueueAddr
	queueAddr string
)

// Exit status to use in case of a usage error
const UsageExitStatus = 7

// NewGlobalFlags creates a new GlobalFlags structure.
func init() {
	gfs = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	gfs.Usage = Usage
	gfs.StringVar(&ConfigFile, "conf", "config.json", "configuration file")
	gfs.StringVar(&queueAddr, "queue", pythia.QueueAddr.String(), "queue address")
}

// AfterParse handles common actions that must be executed after arguments have
// been parsed.
func afterParse() {
	addr, err := pythia.ParseAddr(queueAddr)
	if err != nil {
		UsageError("Invalid address '", queueAddr, "': ", err)
	}
	pythia.QueueAddr = addr
}

// ParseArgs parses command-line arguments. Returns the non-flag arguments.
func parseArgs(args []string) []string {
	if err := gfs.Parse(args); err != nil {
		UsageError(err)
	}
	afterParse()
	return gfs.Args()
}

// ParseMap parses options in a map
func parseMap(options map[string]string) {
	for k, v := range options {
		if err := gfs.Set(k, v); err != nil {
			UsageError(err)
		}
	}
	afterParse()
}

const usageHeader = `Usage: %[1]s [global options]
       %[1]s [global options] component [options]

Launches the pythia platform with the components specified in the configuration
file (first form) or a specific pythia component (second form).

Available components:
`

const componentUsageHeader = `Usage: %[1]s [global options] %[2]s [options]

%[3]s

Options:
`

// Usage prints a notice about the global use of pythia on the standard error.
func Usage() {
	fmt.Fprintf(os.Stderr, usageHeader, os.Args[0])
	for name, info := range pythia.Components {
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", name, info.Description)
	}
	fmt.Fprintf(os.Stderr, "\nGlobal options:\n")
	gfs.PrintDefaults()
}

// UsageError prints msg using fmt.Fprint, followed by the usage notice, and
// exits with error UsageExitStatus.
func UsageError(msg ...interface{}) {
	fmt.Fprint(os.Stderr, os.Args[0], ": ")
	fmt.Fprint(os.Stderr, msg...)
	fmt.Fprint(os.Stderr, "\n")
	Usage()
	os.Exit(UsageExitStatus)
}

// CreateComponentUsage creates a usage function for a component.
func createComponentUsage(info pythia.ComponentInfo, fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, componentUsageHeader, os.Args[0], info.Name,
			info.Description)
		fs.PrintDefaults()
	}
}

// ParseConfig parses the arguments given in the command line and reads the
// configuration file.
// It sets the global parameters and setups the component to run, which is
// returned.
// The program exits with error code 2 in case of an error.
func ParseConfig() pythia.Component {
	parseArgs(os.Args[1:])
	// Parse config file
	rawcfg, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: unable to read configuration file:", err)
	} else {
		var cfg Config
		if err := json.Unmarshal(rawcfg, &cfg); err != nil {
			UsageError(ConfigFile, ": ", err)
		}
		parseMap(cfg.Global)
		ComponentConfigs = make([]ComponentConfig, len(cfg.Components))
		for i, c := range cfg.Components {
			name, ok := c["component"]
			if !ok {
				UsageError(ConfigFile, ": invalid component at index ", i)
			}
			info, ok := pythia.Components[name]
			if !ok {
				UsageError(ConfigFile, ": unknown component '", name, "'")
			}
			delete(c, "component")
			ComponentConfigs[i].Info = info
			ComponentConfigs[i].Options = c
		}
	}
	// Parse CLI arguments again to let them overwrite options in config file
	args := parseArgs(os.Args[1:])
	if len(args) == 0 {
		return NewMaster()
	}
	name := args[0]
	info, ok := pythia.Components[name]
	if !ok {
		UsageError("Unknown component '", name, "'")
	}
	// Setup component
	component := info.New()
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.Usage = createComponentUsage(info, fs)
	if err := component.Setup(fs, args[1:]); err != nil {
		fmt.Fprint(os.Stderr, os.Args[0], " ", name, ": ", err, "\n")
		fs.Usage()
		os.Exit(UsageExitStatus)
	}
	return component
}

// vim:set sw=4 ts=4 noet:
