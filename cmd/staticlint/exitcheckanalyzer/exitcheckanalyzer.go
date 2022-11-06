// Package exitcheckanalyzer is an analyzer prohibiting the use of a direct call
// to os.Exit in the main function of the main package.
package exitcheckanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit direct call in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// inspect all ast nodes
		ast.Inspect(file, func(node ast.Node) bool {
			switch s := node.(type) {
			// check ast for current file
			case *ast.File:
				// check package name
				if s.Name.Name != "main" {
					return false
				}
			case *ast.FuncDecl:
				// check main declaration
				if s.Name.Name != "main" {
					return false
				}
			case *ast.SelectorExpr:
				// check function name
				if s.Sel.Name == "Exit" {
					pass.Reportf(s.Pos(), "os.Exit is not allowed in main package")
				}
			}
			return true
		})
	}
	return nil, nil
}
