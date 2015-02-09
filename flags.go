package main

import (
	"fmt"
	flag "github.com/docker/docker/pkg/mflag"
	"os"
)

var (
	helpPtr       = flag.Bool([]string{"h", "-help"}, false, "Print usage")
	configFilePtr = flag.String([]string{"c", "-config"}, "", "Configuration file")
	versionPtr    = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	numCPU        = flag.Int([]string{"n", "-numcpu"}, 0, "Number of CPUs to use")
	cpuProfilePtr = flag.String([]string{"cp", "-cpuprofile"}, "", "Write cpu profiler results to a given file")
	memProfilePtr = flag.String([]string{"mp", "-memprofile"}, "", "Write heap profile to a given file")
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "Usage: gollum [OPTIONS]\n\nGollum - A n:m message multiplexer.\n\nOptions:\n")

		flag.CommandLine.SetOutput(os.Stdout)

		flag.PrintDefaults()
		fmt.Fprintf(os.Stdout, "\n")
	}
}