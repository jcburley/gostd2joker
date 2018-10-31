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

   Sample routines include (from 'net' package):
     - lookupMX
     - queryEscape
   E.g.:
     ./gostd2joker --dir $PWD/tests 2>&1 | grep -C20 lookupMX

*/

var fset *token.FileSet
var list bool
var dump bool
var verbose bool
var receivers int

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
			s += ", "
		}
		s += n.Name
	}
	return s
}

func fieldListAsString(fl *FieldList) string {
	if fl == nil {
		return ""
	}
	s := ""
	for i, f := range fl.List {
		if i > 0 {
			s += ", "
		}
		if f.Names == nil {
			s += "_"
		} else {
			s += paramNamesAsString(f.Names)
		}
		s += " " + exprAsString(f.Type)
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

func commentGroupInQuotes(doc *CommentGroup) string {
	if doc == nil || doc.Text() == "" {
		return ""
	}
	return `  "` + strings.Trim(doc.Text(), " \t\n") + `"
`
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
				v.Name.Name,
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

type funcInfo struct {
	fd *FuncDecl
	pkg string
}

var functions = map[string]*funcInfo {}
var DUPLICATEFUNCTION = &FuncDecl {}

func processFuncDecl(pkg string, filename string, f *File, fn *FuncDecl) {
	if (dump) {
		Print(fset, fn)
	}
	fname := pkg + "." + fn.Name.Name
	if v, ok := functions[fname]; ok {
		if v.fd != DUPLICATEFUNCTION {
			fmt.Fprintf(os.Stderr, "already seen function %s in %s, yet again in %s\n",
				fname, v.fd, filename)
			fn = DUPLICATEFUNCTION
		}
	}
	functions[fname] = &funcInfo{fn, pkg}
}

var types = map[string]string {}

func processTypeSpec(pkg string, filename string, f *File, ts *TypeSpec) {
	if (dump) {
		Print(fset, ts)
	}
	typename := pkg + "." + ts.Name.Name
	if v, ok := types[typename]; ok {
		if v != "DUPLICATE" {
			fmt.Fprintf(os.Stderr, "already seen type %s in %s, yet again in %s\n",
				typename, v, filename)
			filename = "DUPLICATE"
		}
	}
	types[typename] = filename
}

func processTypeSpecs(pkg string, filename string, f *File, tss []Spec) {
	for _, spec := range tss {
		ts := spec.(*TypeSpec)
		if unicode.IsLower(rune(ts.Name.Name[0])) {
			continue  // Skipping non-exported functions
		}
		processTypeSpec(pkg, filename, f, ts)
	}
}

func processDecls(pkg string, filename string, f *File) {
	for _, s := range f.Decls {
		switch v := s.(type) {
		case *FuncDecl:
			rcv := v.Recv // *FieldList of receivers or nil (functions)
			if rcv != nil {
				receivers += 1
				continue  // Skipping these for now
			}
			if unicode.IsLower(rune(v.Name.Name[0])) {
				continue  // Skipping non-exported functions
			}
			processFuncDecl(pkg, filename, f, v)
		case *GenDecl:
			if v.Tok != token.TYPE {
				continue
			}
			processTypeSpecs(pkg, filename, f, v.Specs)
		default:
			panic("unrecognized Decl type " + fmt.Sprintf("%T", v) + " at: " + fmt.Sprintf("%v", v))
		}
	}
}

func processPackage(pkg string, p *Package) {
	if verbose {
		fmt.Printf("Processing package=%s:\n", pkg)
	}
	for filename, f := range p.Files {
		processDecls(pkg, filename, f)
	}
}

func processDir(d string, path string, mode parser.Mode) error {
	if verbose {
		fmt.Printf("Processing dirname=%s dump=%t:\n", d, dump)
	}

	pkgs, err := parser.ParseDir(fset, path,
		// Walk only *.go files that meet default (target) build constraints, e.g. per "// build ..."
		func (info os.FileInfo) bool { b, e := build.Default.MatchFile(path, info.Name()); return b && e == nil },
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
		basename := filepath.Base(path)
		for k, v := range pkgs {
			if k != basename && k != basename + "_test" {
				if verbose {
					fmt.Printf("NOTICE: Package %s is defined in %s -- ignored\n", k, path)
				}
			} else {
				if verbose {
					fmt.Printf("Package %s:\n", k)
				}
				processPackage(k, v)
//				processPackage(strings.Replace(path, d + "/", "", 1) + "/" + k, v)
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
				if verbose {
					fmt.Printf("Excluding %s\n", path)
				}
				return filepath.SkipDir
			}
			if info.IsDir() {
				if verbose {
					fmt.Printf("Walking from %s to %s\n", d, path)
				}
				return processDir(d, path, mode)
			}
			return nil // not a directory
		})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while walking %s: %v\n", d, err)
		return err
	}

	return err
}

func exprAsClojure(e Expr) string {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			return "String"
		case "int":
			return "Int"
		default:
			return ""
		}
	default:
		return fmt.Sprintf("ABEND881(unrecognized Expr type %T at: %v)", e, e)
	}
}

func exprAsGo(e Expr) string {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			return "string"
		case "int":
			return "int"
		default:
			return "Object"
		}
	default:
		return fmt.Sprintf("ABEND881(unrecognized Expr type %T at: %v)", e, e)
	}
}

func paramNameAsClojure(name *Ident) string {
	return name.Name
}

func fieldListAsClojure(fl *FieldList) string {
	if fl == nil {
		return ""
	}
	var s string
	for _, f := range fl.List {
		cltype := exprAsClojure(f.Type)
		for _, p := range f.Names {
			if s != "" {
				s += ", "
			}
			if cltype != "" {
				s += "^" + cltype + " "
			}
			if p == nil {
				s += "_"
			} else {
				s += paramNameAsClojure(p)
			}
		}
	}
	return s
}

func fieldListToGo(fl *FieldList) string {
	s := ""
	for _, f := range fl.List {
		for _, p := range f.Names {
			if s != "" {
				s += ", "
			}
			if p == nil {
				s += "ABEND922"
			} else {
				s += p.Name
			}
		}
	}
	return s
}

func funcNameAsGoPrivate(f string) string {
	return strings.ToLower(f[0:1]) + f[1:]
}

func paramNameAsGo(p string) string {
	return p
}

func paramListAsGo(fl *FieldList) string {
	s := ""
	for _, f := range fl.List {
		gotype := exprAsGo(f.Type)
		for _, p := range f.Names {
			if s != "" {
				s += ", "
			}
			if p == nil {
				s += "ABEND712"
			} else {
				s += paramNameAsGo(p.Name)
			}
			if gotype != "" {
				s += " " + gotype
			}
		}
	}
	return s
}

func typeAsGo(fl *FieldList) string {
	if fl == nil || fl.List == nil || len(fl.List) < 1 {
		return ""
	}
	if len(fl.List) > 1 {
		return "Object"
	}
	return exprAsGo(fl.List[0].Type)
}

func resultsAsGo(fl *FieldList) string {
	if fl == nil {
		return ""
	}
	s := ""
	arg := 0
	for _, rl := range fl.List {
		if rl.Names == nil {
			arg += 1
			if (arg > 1) {
				s += ", "
			}
			s += fmt.Sprintf("arg_%d", arg)
		} else {
			for _, r := range rl.Names {
				arg += 1
				if (arg > 1) {
					s += ", "
				}
				if r == nil || r.Name == "" {
					s += fmt.Sprintf("arg_%d", arg)
				} else {
					s += r.Name
				}
			}
		}
	}
	return s
}

func argsAsGo(p *FieldList) string {
	s := ""
	for _, f := range p.List {
		for _, p := range f.Names {
			if s != "" {
				s += ", "
			}
			if p == nil {
				s += "ABEND712"
			} else {
				s += paramNameAsGo(p.Name)
			}
		}
	}
	return s
}

func bodyAsGo(pkg string, f *FuncDecl) string {
	callStr := resultsAsGo(f.Type.Results) + " := " + pkg + "." + f.Name.Name + "(" + argsAsGo(f.Type.Params) + ")"

	callStr += "\n...ABEND: TODO..."

	return "\t" + strings.Replace(callStr, "\n", "\n\t", -1)
}

func namedTypeAsClojure(pkg string, t string) string {
	qt := pkg + "." + t
	if v, ok := types[qt]; ok {
		return fmt.Sprintf("ABEND000(have %s for %s)", v, qt)
	} else {
		return fmt.Sprintf("ABEND042(cannot find typename %s)", qt)
	}
}

func typeAsClojure(pkg string, e Expr) string {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			return "String"
		case "int":
			return "Int"
		case "error":
			return "Error"
		default:
			return namedTypeAsClojure(pkg, v.Name)
		}
	case *ArrayType:
		return "[" + typeAsClojure(pkg, v.Elt) + "]"
	case *StarExpr:
		return typeAsClojure(pkg, v.X)
	default:
		return fmt.Sprintf("ABEND881(unrecognized Expr type %T at: %v)", e, e)
	}
}

func returnTypeAsClojure(pkg string, fl *FieldList) string {
	if fl == nil || fl.List == nil {
		return ""
	}
	var s string
	multiple := false
	for _, f := range fl.List {
		cltype := typeAsClojure(pkg, f.Type)
		if f.Names == nil {
			if s != "" {
				s += " "
				multiple = true
			}
			s += cltype
		}
		for _, p := range f.Names {
			if s != "" {
				s += " "
				multiple = true
			}
			if p == nil {
				s += "_"
			} else {
				s += paramNameAsClojure(p)
			}
		}
	}
	if multiple {
		return "[" + s + "]"
	}
	return s
}

func jokerReturnType(pkg string, f *FuncDecl) string {
	s := returnTypeAsClojure(pkg, f.Type.Results)
	if s != "" {
		return "^" + s
	}
	return s
}

var jokerCode = map[string]map[string]string {}
var goCode = map[string]map[string]string {}

func emitFunction(f string, fn *funcInfo) {
	d := fn.fd
	pkg := filepath.Base(fn.pkg)
	jfmt := `
(defn %s%s
%s  {:added "1.0"
   :go "%s(%s)"}
  [%s])
`
	goFname := funcNameAsGoPrivate(d.Name.Name)
	jokerType := jokerReturnType(pkg, d)
	if jokerType != "" {
		jokerType += " "
	}
	jokerFn := fmt.Sprintf(jfmt, jokerType, d.Name.Name, commentGroupInQuotes(d.Doc),
		goFname, fieldListToGo(d.Type.Params),
		fieldListAsClojure(d.Type.Params))

	gfmt := `
func %s(%s) %s {
%s
}
`

	gofn := ""
	if jokerType == "" {
		gofn = fmt.Sprintf(gfmt, goFname, paramListAsGo(d.Type.Params), typeAsGo(d.Type.Results),
			bodyAsGo(pkg, d))
	}

	if strings.Contains(jokerFn, "ABEND") || strings.Contains(gofn, "ABEND") {
		jokerFn = strings.Replace(jokerFn, "\n", "\n;; ", -1)
		gofn = strings.Replace(gofn, "\n", "\n// ", -1)
	}

	if _, ok := jokerCode[pkg]; !ok {
		jokerCode[pkg] = map[string]string {}
	}
	jokerCode[pkg][d.Name.Name] = jokerFn

	if gofn != "" {
		if _, ok := goCode[pkg]; !ok {
			goCode[pkg] = map[string]string {}
		}
		goCode[pkg][d.Name.Name] = gofn
	}
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
		for t, v := range types {
			if verbose {
				fmt.Printf("TYPE %s in %s\n", t, v)
			}
		}
		for f, v := range functions {
			if v.fd == DUPLICATEFUNCTION {
				continue
			}
			if verbose {
				fmt.Printf("FUNC %s in %v\n", f, v)
			}
			emitFunction(f, v)
		}
		for p, v := range jokerCode {
			for f, w := range v {
				fmt.Printf("FUNC %s.%s has: %v\n", p, f, w)
			}
		}
		for p, v := range goCode {
			for f, w := range v {
				fmt.Printf("FUNC %s.%s has: %v\n", p, f, w)
			}
		}
		if verbose {
			fmt.Printf("Totals: types=%d functions=%d receivers=%d\n",
				len(types), len(functions), receivers)
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
