package main

import (
	"fmt"
	. "go/ast"
	"go/parser"
	"go/token"
	"os"
	"unicode"
)

func exprAsString(e Expr) string {
	switch v := e.(type) {
	case *Ident:
		return v.Name
	case *ArrayType:
		return "[]" + exprAsString(v.Elt)
	case *StarExpr:
		return "*" + exprAsString(v.X)
	default:
		panic("unrecognized Expr type")
	}
	return " huh?"
}

func paramNamesAsString(names []*Ident) string {
	s := ""
	for i, n := range names {
		if i > 0 {
			s = s + ", "
		}
		s = s + n.Name
	}
	return s
}

func fieldListAsString(fl *FieldList) string {
	if fl == nil {
		return ""
	}
	var s string
	for i, f := range fl.List {
		if i > 0 {
			s = s + ", "
		}
		if f.Names == nil {
			s = s + "_"
		} else {
			s = s + paramNamesAsString(f.Names)
		}
		s = s + " " + exprAsString(f.Type)
	}
	return s
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	src := `package foo

import (
	"fmt"
	"time"
)

var x Int

func skipMe() {  // lowercase first letter is not exported
	fmt.Println(time.Now())
}

func SomeConverter(i int, s string) (string, error) {
}

func Xyzzy() {
}

func (int) SkipMe() string {  // has a receiver, so not currently handled
}

// LookupMX returns the DNS MX records for the given domain name sorted by preference.
func LookupMX(name string) ([]*MX, error) {
	return DefaultResolver.lookupMX(context.Background(), name)
}

// LookupMX returns the DNS MX records for the given domain name sorted by preference.
func (r *Resolver) LookupMX(ctx context.Context, name string) ([]*MX, error) {
	return r.lookupMX(ctx, name)
}

`

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, /* Or call parser.ParseDir? Also: parser.ImportsOnly, parser.ParseComments ? See https://golang.org/pkg/go/parser/ */ 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--dump" {
		Print(fset, f)
		os.Exit(0)
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	// Now print the decls.
	for _, s := range f.Decls {
		if fn, ok := s.(*FuncDecl); ok {
			rcv := fn.Recv // *FieldList of receivers or nil (functions)
			if rcv != nil {
				continue  // Skipping these for now
			}
			if unicode.IsLower(rune(fn.Name.Name[0])) {
				continue  // Skipping non-exported functions
			}
			typ := fn.Type // *FuncType of signature: params, results, and position of "func" keyword
			fmt.Printf("%s(%s) => %s\n", fn.Name, fieldListAsString(typ.Params), fieldListAsString(typ.Results))
		}
	}
}
