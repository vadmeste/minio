package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	flag.Parse()

	fmt.Println(flag.Arg(0))

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, flag.Arg(0), nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("# List of metrics reported cluster wide")
	fmt.Println("")
	fmt.Println("Each metric includes a label for the server that calculated the metric.")
	fmt.Println("Each metric has a label for the server that generated the metric.")
	fmt.Println("")
	fmt.Println("These metrics can be from any MinIO server once per collection.")
	fmt.Println("")
	fmt.Println("| Name | Description |")
	fmt.Println("|:-----|:------------|")

	// Print the imports from the file's AST.
	for _, c := range f.Comments {
		comment := (*(*c).List[0]).Text
		if !strings.HasPrefix(comment, "//mprom:") {
			continue
		}
		doc := strings.TrimPrefix(comment, "//mprom:")
		s := strings.SplitN(doc, ",", 1)
		if len(s) != 2 {
			log.Fatalln("ERR: found unrecognized mprom doc")
		}

		fmt.Printf("| %s | %s |\n", s[0], s[1])
	}

}
