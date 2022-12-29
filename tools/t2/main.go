package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
)

func main() {
	flag.Parse()

	// Parse src but stop after processing the imports.
	file, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", file, 0)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Store the most specific function declaration
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Type.Results != nil {
			for _, l := range funcDecl.Type.Results.List {
				if id, ok := l.Type.(*ast.Ident); ok && id.Name == "MetricDescription" {
					fmt.Println(string(file[n.Pos()-1 : n.End()]))
				}
			}
		}
		return true
	})
}
