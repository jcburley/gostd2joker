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

var fset *token.FileSet
var dump bool

func chanDirAsString(dir ChanDir) string {
	switch dir {
	case SEND:
		return "chan->"
	case RECV:
		return "<-chan"
	default:
		return "???chan???"
	}
}

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
	case *FuncType:
		return "func(" + fieldListAsString(v.Params) + ")=>" + fieldListAsString(v.Results)
	case *InterfaceType:
		return "interface{" + fieldListAsString(v.Methods) + "}"
	case *Ellipsis:
		return "..." + exprAsString(v.Elt)
	case *MapType:
		return "map[" + exprAsString(v.Key) + "]" + exprAsString(v.Value)
	case *ChanType:
		return chanDirAsString(v.Dir) + " " + exprAsString(v.Value)
	case *StructType:
		return "struct{" + fieldListAsString(v.Fields) + " }"
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

func printTypeSpecs(tss []Spec) {
	for _, spec := range tss {
		ts := spec.(*TypeSpec)
		if unicode.IsLower(rune(ts.Name.Name[0])) {
			continue  // Skipping non-exported functions
		}
		if (dump) {
			Print(fset, ts)
		}
		fmt.Printf("%sTYPE %s %s\n",
			commentGroupAsString(ts.Doc),
			ts.Name.Name,
			exprAsString(ts.Type))
	}
}

func commentGroupAsString(doc *CommentGroup) string {
	if doc == nil {
		return "\n"
	}
	dt := doc.Text()
	nl := strings.Count(dt, "\n")
	if nl > 1 {
		return "\n/* " + strings.Replace(dt, "\n", "\n   ", nl - 1) + "*/\n"
	} else {
		return "\n// " + dt
	}
}

func printDecls(f *File) {
	for _, s := range f.Decls {
		switch v := s.(type) {
		case *FuncDecl:
			rcv := v.Recv // *FieldList of receivers or nil (functions)
			if rcv != nil {
				continue  // Skipping these for now
			}
			if unicode.IsLower(rune(v.Name.Name[0])) {
				continue  // Skipping non-exported functions
			}
			if (dump) {
				Print(fset, v)
			}
			typ := v.Type // *FuncType of signature: params, results, and position of "func" keyword
			fmt.Printf("%s%s(%s) => (%s)\n",
				commentGroupAsString(v.Doc),
				v.Name,
				fieldListAsString(typ.Params),
				fieldListAsString(typ.Results))
		case *GenDecl:
			if v.Tok != token.TYPE {
				continue
			}
			printTypeSpecs(v.Specs)
		default:
			panic("unrecognized Decl type " + fmt.Sprintf("%T", v) + " at: " + fmt.Sprintf("%v", v))
		}
	}
}

func printPackage(p *Package) {
	for n, f := range p.Files {
		fmt.Printf("File %s:\n", n)
		printDecls(f)
	}
}

func processDir(d string, mode parser.Mode) error {
	fmt.Printf("Processing dirname=%s dump=%t:\n", d, dump)

	pkgs, err := parser.ParseDir(fset, d, nil, mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if (dump) {
		for k, v := range pkgs {
			fmt.Printf("Package %s:\n", k)
			printPackage(v)
			fmt.Println("")
		}
	} else {
		basename := filepath.Base(d)
		for k, v := range pkgs {
			if k != basename && k != basename + "_test" {
//				fmt.Printf("NOTICE: Package %s is defined in %s -- ignored\n", k, d)
			} else {
				fmt.Printf("Package %s:\n", k)
				printPackage(v)
			}
		}
	}

	return nil
}

var excludeDirs = map[string]bool {
	"builtin": true,
	"cmd": true,
	"internal": true, // look into this later?
	"testdata": true,
}


func walkDirs(d string, mode parser.Mode) error {
	err := filepath.Walk(d,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s due to: %v\n", path, err)
				return err
			}
			if path == d {
				return nil // skip (implicit) "."
			}
			if excludeDirs[filepath.Base(path)] {
//				fmt.Printf("Excluding %s\n", path)
				return filepath.SkipDir
			}
			if info.IsDir() {
//				fmt.Printf("From %s to %s\n", d, path)
				return processDir(path, mode)
			}
			return nil // not a directory
		})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while walking %s: %v\n", d, err)
		return err
	}

	return err
}

func notOption(arg string) bool {
	return arg == "-" || !strings.HasPrefix(arg, "-")
}

func main() {
	fset = token.NewFileSet() // positions are relative to fset
	dump = false

	length := len(os.Args)
	filename := ""
	dir := ""
	var mode parser.Mode = parser.ParseComments /* Also: parser.ImportsOnly, parser.ParseComments ? See https://golang.org/pkg/go/parser/ */

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
		err := walkDirs(dir, mode)
		if err != nil {
			panic("Error walking directory " + dir + ": " + fmt.Sprintf("%v", err))
		}
		os.Exit(0)
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
