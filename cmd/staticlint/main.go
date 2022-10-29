/* Package staticlint defines a multi-checker for static analysis.
   It implements tools/go/analysis/multichecker interface.

   The multichecker contains:
	1) standard static analyzers from the golang.org/x/tools/go/analysis/passes package.
	2) all analyzers of the SA class of the staticcheck package.
	3) all analyzers of the S class of the simple package.
	4) all analyzers of the ST class of the stylecheck package.
	5) custom analyzer exitcheckanalyzer to check call of os.Exit in main package.
    6) open source go-critic analyzer.
    7) open source asciicheck analyzer.

   How to use:
	1) go build from the cmd/staticlint
    2) run from the root folder using command: .\cmd\staticlint\staticlint.exe ./...
*/
package main

import (
	"encoding/json"
	"fmt"
	"go-developer-course-shortener/cmd/staticlint/exitcheckanalyzer"

	"os"
	"path/filepath"
	"strings"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/tdakkota/asciicheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config configuration file name
const Config = `config/config.json`

// ConfigData structure for configuration
type ConfigData struct {
	Staticcheck []string
	Simple      []string
	Stylecheck  []string
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	// print current configuration
	fmt.Printf("%+v\n\n", cfg)

	mychecks := []*analysis.Analyzer{
		exitcheckanalyzer.ExitCheckAnalyzer, // custom analyzer to check os.Exit in main package
		gocritic.Analyzer,                   // go-critic analyzer
		asciicheck.NewAnalyzer(),            // ascii check analyzer
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}

	// set static analyzers from config file
	for _, v := range staticcheck.Analyzers {
		for _, c := range cfg.Staticcheck {
			if strings.HasPrefix(v.Name, c) {
				mychecks = append(mychecks, v)
			}
		}
	}

	// set simple analyzers from config file
	for _, v := range simple.Analyzers {
		for _, c := range cfg.Simple {
			if strings.HasPrefix(v.Name, c) {
				mychecks = append(mychecks, v)
			}
		}
	}

	// set style analyzers from config file
	for _, v := range stylecheck.Analyzers {
		for _, c := range cfg.Stylecheck {
			if strings.HasPrefix(v.Name, c) {
				mychecks = append(mychecks, v)
			}
		}
	}

	fmt.Println("Staticlint started")
	fmt.Println("The following analyzers are included:\n", mychecks)

	multichecker.Main(
		mychecks...,
	)

}
