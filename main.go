// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main // import "github.com/minio/minio"

import (
	// MUST be first import.

	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/minio/minio/cmd"
	minio "github.com/minio/minio/cmd"
	_ "github.com/minio/minio/internal/init"
)

func main() {

	write := false // true
	source := "cmd/metrics-v2-def.go"

	// Parse src but stop after processing the imports.
	file, err := ioutil.ReadFile(source)
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

	if write {
		fmt.Println("package cmd")
		fmt.Println("type My struct{}")
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
					if write {
						x := bytes.Index(file[n.Pos():n.End()], []byte{'\n'})
						if x >= 0 {
							fmt.Printf("func (m My) %s() MetricDescription {\n", strings.ToUpper(string(funcDecl.Name.Name[0]))+funcDecl.Name.Name[1:])
							fmt.Println(string(file[int(n.Pos())+x : n.End()]))
						}
					} else {
						x := bytes.Index(file[n.Pos():n.End()], []byte{'\n'})
						if x >= 0 {
							m := cmd.My{}
							meth := reflect.ValueOf(m).MethodByName(strings.ToUpper(string(funcDecl.Name.Name[0])) + funcDecl.Name.Name[1:])
							ii := meth.Call(nil)
							// fmt.Println(reflect.TypeOf(ii))
							// fmt.Println(reflect.TypeOf(ii[0]))
							md, _ := ii[0].Interface().(cmd.MetricDescription)
							fmt.Printf("//prom: %s_%s_%s, %s\n", md.Namespace, md.Subsystem, md.Name, md.Help)
							fmt.Printf("func %s() MetricDescription {", funcDecl.Name.Name)
							fmt.Println(string(file[int(n.Pos())+x : n.End()]))
						}

						/*
							fmt.Println(funcDecl.Name.Name)
							var fn func() MetricDescription
							err := GetFunc(&fn, "github.com/minio/minio."+funcDecl.Name.Name)
							if err != nil {
								// Handle errors if you care about name possibly being invalid.
							}
							fmt.Println(err)
						*/
					}
				}
			}
		}
		return true
	})

	return
	minio.Main(os.Args)
}
