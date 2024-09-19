package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jondot/goweight/pkg"
)

var (
	jsonOutput bool
	buildTags  string
	workDir    string
	ldflags    string
	csvOutput  bool
)

func init() {
	flag.BoolVar(&jsonOutput, "json", false, "Output json format")
	flag.BoolVar(&jsonOutput, "j", false, "Output json format (shorthand)")
	flag.StringVar(&buildTags, "tags", "", "Build tags")
	flag.StringVar(&workDir, "workdir", "", "Work directory")
	flag.StringVar(&workDir, "w", "", "Work directory (shorthand)")
	flag.StringVar(&ldflags, "ldflags", "", "arguments to pass on each go tool link invocation. Default ''")
	flag.BoolVar(&csvOutput, "csv", false, "Output csv format. Default generated file is the current directory 'out.csv'")
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

	} else if csvOutput {
		file, err := os.Create("out.csv")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		writer.Write([]string{"Size", "Module"})

		// Write data
		for _, module := range modules {
			writer.Write([]string{module.SizeHuman, module.Name})
		}

	} else {
		fmt.Printf("%8s %s\n", "Size", "Module")
		for _, module := range modules {
			fmt.Printf("%8s %s\n", module.SizeHuman, module.Name)
		}
	}
}
