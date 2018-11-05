package main

import (
	"bufio"
	"fmt"
	. "go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"unicode"
)

const VERSION = "0.1"

func check(e error) {
    if e != nil {
        panic(e)
    }
}

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
var populateDir string

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

func whereAt(p token.Pos) string {
	return fmt.Sprintf("%s", fset.Position(p).String())
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
		panic(fmt.Sprintf("unrecognized Expr type %T at: %s", e, whereAt(v.Pos())))
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
	return `  ` + strings.Trim(strconv.Quote(doc.Text()), " \t\n") + "\n"
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
			panic(fmt.Sprintf("unrecognized Decl type %T at: %s", v, whereAt(v.Pos())))
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
	filename string
}

var functions = map[string]*funcInfo {}
var DUPLICATEFUNCTION = &FuncDecl {}

var alreadySeen = []string {}

func processFuncDecl(pkg string, filename string, f *File, fn *FuncDecl) {
	if (dump) {
		Print(fset, fn)
	}
	fname := pkg + "." + fn.Name.Name
	if v, ok := functions[fname]; ok {
		if v.fd != DUPLICATEFUNCTION {
			alreadySeen = append(alreadySeen,
				fmt.Sprintf("NOTE: Already seen function %s in %s, yet again in %s",
					fname, v.filename, filename))
			fn = DUPLICATEFUNCTION
		}
	}
	functions[fname] = &funcInfo{fn, pkg, filename}
}

type typeInfo struct {
	td *TypeSpec
	file string
}

type typeInfoArray []*typeInfo

/* Go apparently doesn't support/allow 'interface{}' as the value (or
/* key) of a map such that any arbitrary type can be substituted at
/* run time, so there are several of these nearly-identical functions
/* sprinkled through this code. Still get some reuse out of some of
/* them, and it's still easier to maintain these copies than if the
/* body of these were to be included at each call point.... */
func sortedTypeInfoMap(m map[string]typeInfoArray, f func(k string, v typeInfoArray)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}

var types = map[string]typeInfoArray {}

func processTypeSpec(pkg string, filename string, f *File, ts *TypeSpec) {
	if (dump) {
		Print(fset, ts)
	}
	typename := pkg + "." + ts.Name.Name
	var candidates typeInfoArray
	if candidates, ok := types[typename]; ok {
		for _, c := range candidates {
			if c.file == filename {
				panic(fmt.Sprintf("type %s defined twice in file %s", typename, filename))
			}
		}
	} else {
		candidates = typeInfoArray {}
	}
	candidates = append(candidates, &typeInfo{ts, filename})
	types[typename] = candidates
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
			panic(fmt.Sprintf("unrecognized Decl type %T at: %s", v, whereAt(v.Pos())))
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
		fmt.Printf("Processing sourceDir=%s dump=%t:\n", d, dump)
	}

	pkgs, err := parser.ParseDir(fset, path,
		// Walk only *.go files that meet default (target) build constraints, e.g. per "// build ..."
		func (info os.FileInfo) bool {
			if strings.HasSuffix(info.Name(), "_test.go") {
				return false
			}
			b, e := build.Default.MatchFile(path, info.Name());
			return b && e == nil
		},
		mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if list {
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
	target, err := filepath.EvalSymlinks(d)
	check(err)
	err = filepath.Walk(target,
		func(path string, info os.FileInfo, err error) error {
			rel := strings.Replace(path, target, d, 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s due to: %v\n", rel, err)
				return err
			}
			if rel == d {
				return nil // skip (implicit) "."
			}
			if excludeDirs[filepath.Base(rel)] {
				if verbose {
					fmt.Printf("Excluding %s\n", rel)
				}
				return filepath.SkipDir
			}
			if info.IsDir() {
				if verbose {
					fmt.Printf("Walking from %s to %s\n", d, rel)
				}
				return processDir(d, rel, mode)
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
		case "int", "uint", "int16", "uint16":
			return "Int"
		default:
			return ""
		}
	default:
		return fmt.Sprintf("ABEND881(unrecognized Expr type %T at: %s)", e, whereAt(e.Pos()))
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
		return fmt.Sprintf("ABEND882(unrecognized Expr type %T at: %s)", e, whereAt(e.Pos()))
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
		return typeAsClojure(pkg, v[0].td.Type)
	} else {
		return fmt.Sprintf("ABEND042(cannot find typename %s)", qt)
	}
}

// E.g.: {:host ^String "Host" :pref ^Int "Pref"}
func structAsClojure(pkg string, fl *FieldList) string {
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
			if p == nil {
				s += "_"
			} else {
				s += ":" + strings.ToLower(p.Name)
			}
			if cltype != "" {
				s += " ^" + cltype
			}
			if p == nil {
				s += " _"
			} else {
				s += " " + p.Name
			}
		}
	}
	return s
}

func typeAsClojure(pkg string, e Expr) string {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			return "String"
		case "int", "int16", "uint", "uint16":
			return "Int"
		case "error":
			return "Error"
		default:
			return namedTypeAsClojure(pkg, v.Name)
		}
/*
	case *ArrayType:
		return "[" + typeAsClojure(pkg, v.Elt) + "]"
*/
	case *StarExpr:
		return typeAsClojure(pkg, v.X)
/*
	case *StructType:
		return "{" + structAsClojure(pkg, v.Fields) + "}"
*/
	default:
		return fmt.Sprintf("ABEND883(unrecognized Expr type %T at: %s)", e, whereAt(e.Pos()))
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

/* Map package names to maps of filenames to code strings. */

type codeInfo map[string]string

var jokerCode = map[string]codeInfo {}
var goCode = map[string]codeInfo {}

func sortedPackageMap(m map[string]codeInfo, f func(k string, v codeInfo)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}

func sortedCodeMap(m codeInfo, f func(k string, v string)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}

var nonEmptyLineRegexp *regexp.Regexp

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
		jokerFn = nonEmptyLineRegexp.ReplaceAllString(jokerFn, `;; $1`)
		gofn = nonEmptyLineRegexp.ReplaceAllString(gofn, `// $1`)
	}

	if _, ok := jokerCode[pkg]; !ok {
		jokerCode[pkg] = codeInfo {}
	}
	jokerCode[pkg][d.Name.Name] = jokerFn

	if gofn != "" {
		if _, ok := goCode[pkg]; !ok {
			goCode[pkg] = codeInfo {}
		}
		goCode[pkg][d.Name.Name] = gofn
	}
}

func notOption(arg string) bool {
	return arg == "-" || !strings.HasPrefix(arg, "-")
}

func usage() {
	fmt.Print(`
Usage: gostd2joker options...

Options:
  --source <go-source-dir-name>  # Location of Go source tree's src/ subdirectory
  --populate <joker-std-subdir>  # Where to write the joker.go.* files (usually .../joker/std/go/)
  --overwrite                    # Overwrite any existing <joker-std-subdir> files, leaving existing files intact
  --replace                      # 'rm -fr <joker-std-subdir>' before creating <joker-std-subdir>
  --fresh                        # (Default) Refuse to overwrite existing <joker-std-subdir> directory
  --verbose, -v                  # Print info on what's going on
  --list                         # List packages, as each walked directory is processed, instead of processing them
  --dump                         # Use go's AST dump API on pertinent elements (functions, types, etc.)
  --help, -h                     # Print this information

If <joker-std-subdir> is not specified, no Go nor Clojure source files
(nor any other files nor directories) are created, effecting a sort of
"dry run".
`)
	os.Exit(0)
}

func main() {
	fset = token.NewFileSet() // positions are relative to fset
	dump = false

	length := len(os.Args)
	sourceDir := ""
	replace := false
	overwrite := false

	var mode parser.Mode = parser.ParseComments /* Also: parser.ImportsOnly, parser.ParseComments ? See https://golang.org/pkg/go/parser/ */

	for i := 1; i < length; i++ { // shift
		a := os.Args[i]
		if a[0] == "-"[0] {
			switch a {
			case "--help", "-h":
				usage()
			case "--version", "-V":
				fmt.Printf("%s version %s\n", os.Args[0], VERSION)
			case "--populate":
				if populateDir != "" {
					panic("cannot specify --populate <joker-std-subdir> more than once")
				}
				if i < length-1 && notOption(os.Args[i+1]) {
					i += 1 // shift
					populateDir = os.Args[i]
				} else {
					panic("missing path after --populate option")
				}
			case "--dump":
				dump = true
			case "--overwrite":
				overwrite = true
				replace = false
			case "--replace":
				replace = true
				overwrite = false
			case "--fresh":
				replace = false
				overwrite = false
			case "--list":
				list = true
			case "--verbose", "-v":
				verbose = true
			case "--source":
				if sourceDir != "" {
					panic("cannot specify --source <go-source-dir-name> more than once")
				}
				if i < length-1 && notOption(os.Args[i+1]) {
					i += 1 // shift
					sourceDir = os.Args[i]
				} else {
					panic("missing path after --source option")
				}
			default:
				panic("unrecognized option " + a)
			}
		} else {
			panic("extraneous argument(s) starting with: " + a)
		}
	}

	if verbose {
		fmt.Printf("Default context: %v\n", build.Default)
	}

	if sourceDir == "" {
		panic("Must specify --source <go-source-dir-name> option")
	}

	if fi, e := os.Stat(filepath.Join(sourceDir, "go")); e != nil || !fi.IsDir() {
		if m, e := filepath.Glob(filepath.Join(sourceDir, "*.go")); e != nil || m == nil || len(m) == 0 {
			panic(fmt.Sprintf("Does not exist or is not a Go source directory: %s;\n%v", sourceDir, m))
		}
	}

	if populateDir != "" {
		if replace {
			if e := os.RemoveAll(populateDir); e != nil {
				panic(fmt.Sprintf("Unable to effectively 'rm -fr %s'", populateDir))
			}
		}

		if !overwrite {
			var stat syscall.Stat_t
			if e := syscall.Stat(populateDir, &stat); e == nil || e.Error() != "no such file or directory" {
				msg := "already exists"
				if e != nil {
					msg = e.Error()
				}
				panic(fmt.Sprintf("Cannot populate empty directory %s; please 'rm -fr' first, or specify --overwrite or --replace: %s",
					populateDir, msg))
			}
			if e := os.MkdirAll(populateDir, 0777); e != nil {
				panic(fmt.Sprintf("Cannot 'mkdir -p %s': %s", populateDir, e.Error()))
			}
		}
	}

	err := walkDirs(filepath.Join(sourceDir, "."), mode)
	if err != nil {
		panic("Error walking directory " + sourceDir + ": " + fmt.Sprintf("%v", err))
	}

	sort.Strings(alreadySeen)
	for _, a := range alreadySeen {
		fmt.Fprintln(os.Stderr, a)
	}

	if verbose {
		/* Output map in sorted order to stabilize for testing. */
		sortedTypeInfoMap(types,
			func(t string, v typeInfoArray) {
				fmt.Printf("TYPE %s:\n", t)
				for _, ts := range v {
					fmt.Printf("  %s\n", ts.file)
				}
			})
	}

	/* Emit function code snippets in arbitrary/random order. */
	for f, v := range functions {
		if v.fd == DUPLICATEFUNCTION {
			continue
		}
		emitFunction(f, v)
	}

	var out *bufio.Writer
	var unbuf_out *os.File

	sortedPackageMap(jokerCode,
		func(p string, v codeInfo) {
			if populateDir != "" {
				jf := filepath.Join(populateDir, p + ".joke")
				var e error
				unbuf_out, e = os.Create(jf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)
				fmt.Fprintf(out, `;;;; Auto-generated by gostd2joker, do not edit!!

(ns
  ^{:go-imports []
    :doc "Provides a low-level interface to the %s package."}
  go.%s)
`,
					p, p)
			}
			sortedCodeMap(v,
				func(f string, w string) {
					fmt.Printf("JOKER FUNC %s.%s has:%v\n", p, f, w)
					if out != nil {
						out.WriteString(w)
					}
				})
			if out != nil {
				out.Flush()
				unbuf_out.Close()
				out = nil
			}
		})
	sortedPackageMap(goCode,
		func(p string, v codeInfo) {
			if populateDir != "" {
				gf := filepath.Join(populateDir, p, p + "_native.go")
				var e error
				e = os.MkdirAll(filepath.Dir(gf), 0777)
				check(e)
				unbuf_out, e = os.Create(gf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)
				fmt.Fprintf(out, `// Auto-generated by gostd2joker, do not edit!!

package %s

import (
	"%s"
)
`,
					p, p)
			}
			sortedCodeMap(v,
				func(f string, w string) {
					if out != nil {
						out.WriteString(w)
					}
					fmt.Printf("GO FUNC %s.%s has:%v\n", p, f, w)
				})
			if out != nil {
				out.Flush()
				unbuf_out.Close()
				out = nil
			}
		})

	if verbose {
		fmt.Printf("Totals: types=%d functions=%d receivers=%d\n",
			len(types), len(functions), receivers)
	}

	os.Exit(0)
}

func init() {
	p := `(?m)^(.)`
	nonEmptyLineRegexp = regexp.MustCompile(p)
}
