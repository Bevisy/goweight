package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/jondot/goweight/pkg"
)

var (
	version = "dev"
	commit  = "none"
)

var (
	jsonOutput bool
	buildTags  string
	workDir    string
	ldflags    string
)

func init() {
	flag.BoolVar(&jsonOutput, "json", false, "Output json")
	flag.BoolVar(&jsonOutput, "j", false, "Output json (shorthand)")
	flag.StringVar(&buildTags, "tags", "", "Build tags")
	flag.StringVar(&workDir, "workdir", "", "Work directory")
	flag.StringVar(&workDir, "w", "", "Work directory (shorthand)")
	flag.StringVar(&ldflags, "ldflags", "", "arguments to pass on each go tool link invocation. Default ''")
}

func main() {
	flag.Parse()

	weight := pkg.NewGoWeight(workDir)
	if buildTags != "" {
		weight.BuildCmd = append(weight.BuildCmd, "-tags", buildTags)
	}

	if ldflags != "" {
		weight.BuildCmd = append(weight.BuildCmd, "-ldflags", ldflags)
	}

	// Append non-flag arguments to weight.BuildCmd
	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) > 0 {
		weight.BuildCmd = append(weight.BuildCmd, nonFlagArgs...)
	}

	work := weight.BuildCurrent()
	// add "/" suffix to work directory, or zenv root directory will be wrong
	modules := weight.Process(work + "/")

	if jsonOutput {
		m, _ := json.Marshal(modules)
		fmt.Print(string(m))
	} else {
		for _, module := range modules {
			fmt.Printf("%8s %s\n", module.SizeHuman, module.Name)
		}
	}
}
