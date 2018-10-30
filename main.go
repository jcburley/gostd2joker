package main

import (
	"fmt"
	. "go/ast"
	"go/build"
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
var list bool
var dump bool
var verbose bool

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
	for nam, f := range p.Files {
		fmt.Printf("File %s:\n", nam)
		printDecls(f)
	}
}

func processFuncDecl(pkg string, name string, f *File, fn *FuncDecl) {
	if (dump) {
		Print(fset, fn)
	}
}

var types = map[string]bool {}

func processTypeSpec(pkg string, name string, f *File, ts *TypeSpec) {
	if (dump) {
		Print(fset, ts)
	}
	typename := pkg + "." + ts.Name.Name
	if types[typename] {
		panic("already seen type " + typename + ", yet again in " + name)
	}
	types[typename] = true
}

func processTypeSpecs(pkg string, name string, f *File, tss []Spec) {
	for _, spec := range tss {
		ts := spec.(*TypeSpec)
		if unicode.IsLower(rune(ts.Name.Name[0])) {
			continue  // Skipping non-exported functions
		}
		processTypeSpec(pkg, name, f, ts)
	}
}

func processDecls(pkg string, name string, f *File) {
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
			processFuncDecl(pkg, name, f, v)
		case *GenDecl:
			if v.Tok != token.TYPE {
				continue
			}
			processTypeSpecs(pkg, name, f, v.Specs)
		default:
			panic("unrecognized Decl type " + fmt.Sprintf("%T", v) + " at: " + fmt.Sprintf("%v", v))
		}
	}
}

func processPackage(pkg string, p *Package) {
	for name, f := range p.Files {
		processDecls(pkg, name, f)
	}
}

func processDir(d string, mode parser.Mode) error {
	if (verbose) {
		fmt.Printf("Processing dirname=%s dump=%t:\n", d, dump)
	}

	pkgs, err := parser.ParseDir(fset, d,
		// Walk only *.go files that meet default (target) build constraints, e.g. per "// build ..."
		func (info os.FileInfo) bool { b, e := build.Default.MatchFile(d, info.Name()); return b && e != nil },
		mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if (list) {
		for k, v := range pkgs {
			fmt.Printf("Package %s:\n", k)
			printPackage(v)
			fmt.Println("")
		}
	} else {
		basename := filepath.Base(d)
		for k, v := range pkgs {
			if k != basename && k != basename + "_test" {
				if (verbose) {
					fmt.Printf("NOTICE: Package %s is defined in %s -- ignored\n", k, d)
				}
			} else {
				if (verbose) {
					fmt.Printf("Package %s:\n", k)
				}
				processPackage(k, v)
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
				if (verbose) {
					fmt.Printf("Excluding %s\n", path)
				}
				return filepath.SkipDir
			}
			if info.IsDir() {
				if (verbose) {
					fmt.Printf("Walking from %s to %s\n", d, path)
				}
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
			case "--list":
				list = true
			case "--verbose", "-v":
				verbose = true
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

	if verbose {
		fmt.Printf("Default context:\n%v\n", build.Default)
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
		if list {
			os.Exit(0)
		}
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	// Now print the decls.
	printDecls(f)
}
