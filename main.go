package main

import (
	"fmt"
	. "go/ast"
	"go/parser"
	"go/token"
	"strings"
	"os"
	"path/filepath"
	"unicode"
)

/* Want to support e.g.:

     net/dial.go:DialTimeout(network, address string, timeout time.Duration) => _ Conn, _ error

   I.e. a function defined in one package refers to a type defined in
   another (a different directory, even).

*/

func exprAsString(e Expr) string {
	switch v := e.(type) {
	case *Ident:
		return v.Name
	case *ArrayType:
		return "[]" + exprAsString(v.Elt)
	case *StarExpr:
		return "*" + exprAsString(v.X)
	case *SelectorExpr:
		return exprAsString(v.X) + "." + v.Sel.Name
	default:
		panic("unrecognized Expr type " + fmt.Sprintf("%T", e) + " at: " + fmt.Sprintf("%v", e))
	}
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

func printDecls(f *File) {
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

func printPackage(p *Package) {
	for n, f := range p.Files {
		fmt.Printf("File %s:\n", n)
		printDecls(f)
	}
}

func processDir(d string, mode parser.Mode, dump bool) int {
	fmt.Printf("Processing dirname=%s dump=%t:\n", d, dump)

	fset := token.NewFileSet() // positions are relative to fset
	pkgs, err := parser.ParseDir(fset, d, nil, mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if (dump) {
		for k, v := range pkgs {
			fmt.Printf("Package %s:\n", k)
			printPackage(v)
			fmt.Println("")
		}
	} else {
		basename := filepath.Base(d)
		for k, _ := range pkgs {
			if k != basename && k != basename + "_test" {
				fmt.Printf("PROBLEM: Package %s is defined in %s\n", k, d)
			}
		}
	}
	
	return 0
}



func notOption(arg string) bool {
	return arg == "-" || !strings.HasPrefix(arg, "-")
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	length := len(os.Args)
	dump := false
	filename := ""
	dir := ""
	var mode parser.Mode = 0 /* Also: parser.ImportsOnly, parser.ParseComments ? See https://golang.org/pkg/go/parser/ */

	for i := 1; i < length; i++ { // shift
		a := os.Args[i]
		if a[0] == "-"[0] {
			switch a {
			case "--dump":
				dump = true
			case "--dir":
				if filename != "" {
					panic("cannot specify both a filename and the --dir <dirname> option")
				}
				if dir != "" {
					panic("cannot specify --dir <dirname> more than once")
				}
				if i < length-1 && notOption(os.Args[i+1]) {
					i += 1 // shift
					dir = os.Args[i]
				} else {
					panic("missing path after --dir option")
				}
			default:
				panic("unrecognized option " + a)
			}
		} else if filename == "" {
			filename = a
		} else {
			panic("only one filename may be specified on command line: " + a)
		}
	}

	if dir != "" {
		os.Exit(processDir(dir, mode, dump))
	}

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

	f, err := parser.ParseFile(fset, filename,
		func () interface{} { if filename == "" { return src } else { return nil } }(),
		mode)
	if err != nil {
		fmt.Println(err)
		return
	}

	if dump {
		Print(fset, f)
		os.Exit(0)
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	// Now print the decls.
	printDecls(f)
}
