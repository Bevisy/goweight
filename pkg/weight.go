package pkg

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/mattn/go-zglob"
	"github.com/thoas/go-funk"
)

const BIN_TARGET = "goweight-bin-target"

var moduleRegex = regexp.MustCompile("packagefile (.*)=(.*)")

func run(cmd GoWeight) string {
	fmt.Printf("execute: %s\n", strings.Join(cmd.BuildCmd, " "))

	cmdE := exec.Command(cmd.BuildCmd[0], cmd.BuildCmd[1:]...)
	out, err := cmdE.CombinedOutput()
	if err != nil {
		log.Fatal(fmt.Errorf("command: '%s' execute failed: %s : %v", strings.Join(cmd.BuildCmd, " "), out, err))
	}
	os.Remove(BIN_TARGET)
	return string(out)
}

func processModule(line string) *ModuleEntry {
	captures := moduleRegex.FindAllStringSubmatch(line, -1)
	if captures == nil {
		return nil
	}
	path := captures[0][2]
	stat, _ := os.Stat(path)
	sz := uint64(stat.Size())

	return &ModuleEntry{
		Path:      path,
		Name:      captures[0][1],
		Size:      sz,
		SizeHuman: humanize.Bytes(sz),
	}
}

type ModuleEntry struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	SizeHuman string `json:"size_human"`
}
type GoWeight struct {
	BuildCmd []string
	BuildEnv []string
}

func NewGoWeight(workDir string) *GoWeight {
	var err error
	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
	}
	return &GoWeight{
		BuildCmd: []string{"go", "build", "-C", workDir, "-o", BIN_TARGET, "-work", "-a"},
	}
}

// return go build temporary directory
func (g *GoWeight) BuildCurrent() string {
	d := strings.Split(strings.TrimSpace(run(*g)), "\n")[0]
	return strings.Split(strings.TrimSpace(d), "=")[1]
}

func (g *GoWeight) Process(work string) []*ModuleEntry {
	files, err := zglob.Glob(work + "**/importcfg")
	if err != nil {
		log.Fatal(fmt.Errorf("could not open directory %s: %w", work, err))
	}

	allLines := funk.Uniq(funk.FlattenDeep(funk.Map(files, func(file string) []string {
		f, err := os.ReadFile(file)
		if err != nil {
			return []string{}
		}
		return strings.Split(string(f), "\n")
	})))
	modules := funk.Compact(funk.Map(allLines, processModule)).([]*ModuleEntry)
	sort.Slice(modules, func(i, j int) bool { return modules[i].Size > modules[j].Size })

	return modules
}

// ChangePermissions recursively changes the permissions of the directory and its contents to 777
func Perm(path string, mode fs.FileMode) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(path, mode)
	})
}
