// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates contributors.go. It can be invoked by running
// go generate

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var 

func main() {
	fs := token.NewFileSet()
	a, err := parser.ParseFile(fs, "scout.1.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var v visitor
	ast.Walk(v, a)
}

type visitor int

func (v visitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	// fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), node)

	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Doc != nil {
			if strings.Contains(d.Doc.List[0].Text, "gen must") {
				spew.Dump(d)
			}
		}
	}

	return v + 1
}
