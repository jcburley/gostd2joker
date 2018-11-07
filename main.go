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
var dump bool
var verbose bool
var receivers int
var populateDir string

func whereAt(p token.Pos) string {
	return fmt.Sprintf("%s", fset.Position(p).String())
}

func commentGroupInQuotes(doc *CommentGroup, jok, gol string) string {
	var d string
	if doc != nil {
		d = doc.Text()
	}
	if gol != "" {
		if d != "" {
			d = strings.Trim(d, " \t\n") + "\n"
		}
		d += "Go return type: " + gol
	}
	if jok != "" {
		if d != "" {
			d = strings.Trim(d, " \t\n") + "\n"
		}
		d += "Joker return type: " + jok
	}
	return `  ` + strings.Trim(strconv.Quote(d), " \t\n") + "\n"
}

type funcInfo struct {
	fd *FuncDecl
	pkg string
	filename string
}

/* Go apparently doesn't support/allow 'interface{}' as the value (or
/* key) of a map such that any arbitrary type can be substituted at
/* run time, so there are several of these nearly-identical functions
/* sprinkled through this code. Still get some reuse out of some of
/* them, and it's still easier to maintain these copies than if the
/* body of these were to be included at each call point.... */
func sortedFuncInfoMap(m map[string]*funcInfo, f func(k string, v *funcInfo)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
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
	building bool
	built bool
	jok string
	gol string
}

type typeInfoArray []*typeInfo

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
	candidates = append(candidates, &typeInfo{ts, filename, false, false, "", ""})
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

func paramNameAsClojure(n string) string {
	return n
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
				s += paramNameAsClojure(p.Name)
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

func argsAsGo(p *FieldList) string {
	s := ""
	for _, f := range p.List {
		for _, p := range f.Names {
			if s != "" {
				s += ", "
			}
			if p == nil {
				s += "ABEND713"
			} else {
				s += paramNameAsGo(p.Name)
			}
		}
	}
	return s
}

/* The transformation code, below, takes an approach that is new for me.

   Instead of each transformation point having its own transform
   routine(s), as is customary, I'm trying an approach in which the
   transform is driven by the input and multiple outputs are
   generated, where appropriate, for further processing and/or
   insertion into the ultimate transformation points.

   The primary reason for this is that the input is complicated and
   (generally) being supported to a greater extent as enhancements are
   made. I want to maintain coherence among the various transformation
   insertions, so it's less likely that a change made for one
   insertion point (to support a new input form, or modify an existing
   one) won't have corresponding changes made to other forms relying
   on the same essential input, which could lead to coding errors.

   This approach also should make it easier to see how the different
   snippets of code, relating to one particular aspect of the input,
   relate to each other, because the code will be in the same place.

   However, I'm concerned that the resulting code will be too
   complicated for that to be sufficiently helpful. If I was
   proficient in a constraint/unification-based transformation
   language, I'd look at that instead, because it would allow me to
   express that e.g. "func foo(args) (returns) { ...do things with
   args...; call foo in some fashion; ...do things with returns... }"
   not only have specific transformations for each of the variables
   involved, but that they are also constrained in some fashion
   (e.g. whatever names are picked for unnamed 'returns' values are
   the same in both "returns" and "do things with returns"; whatever
   types are involved in both "args" and "returns" are properly
   processed in "do things with args" and "do things with returns",
   respectively; and so on).

   Now that I've refactored the code to achieve this, I'll start
   adding transformations and see how it goes. Might revert to
   old-fashioned use of custom transformation code per point (sharing
   code where appropriate, of course) if it gets too hairy.

 */

func genGoPostNamed(indent, pkg, tmp, t string) (jok, gol string) {
	qt := pkg + "." + t
	if v, ok := types[qt]; ok {
		if v[0].built { // Reuse already-built type info
			jok = v[0].jok
			gol = v[0].gol
			return
		}
		if v[0].building { // Mutually-referring types currently not supported
			jok = fmt.Sprintf("ABEND947(recursive type reference involving %s)", qt)  // TODO: handle these, e.g. http Request/Response
			gol = jok
		} else {
			v[0].building = true
			jok, gol = genGoPostElement(indent, pkg, tmp, v[0].td.Type)
		}
		v[0].jok = jok
		v[0].gol = gol
		v[0].built = true
		return
	} else {
		jok = fmt.Sprintf("ABEND042(cannot find typename %s)", qt)
		return
	}
}

// func tryThis(s string) struct { a int; b string } {
//	return struct { a int; b string }{ 5, "hey" }
// }

// Joker: { :a ^Int, :b ^String }
// Go: struct { a int; b string }
func genGoPostStruct(indent, pkg, tmp string, fl *FieldList) (jok, gol string) {
	if fl == nil {
		jok = "{}"
		gol = "struct {}"
		return
	}
	for _, f := range fl.List {
		joktype, goltype := genGoPostElement(indent, pkg, tmp, f.Type)
		for _, p := range f.Names {
			if jok != "" {
				jok += ", "
			}
			if gol != "" {
				gol += "; "
			}
			if p == nil {
				jok += "_ "
			} else {
				jok += ":" + p.Name + " "
				gol += p.Name + " "
			}
			if joktype != "" {
				jok += "^" + joktype
			}
			if goltype != "" {
				gol += goltype
			}
		}
	}
	jok = "{" + jok + "}"
	gol = "struct {" + gol + "}"
	return
}

func genGoPostElement(indent, pkg, tmp string, e Expr) (jok, gol string) {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			jok = "String"
			gol = "string"
			return
		case "int", "int16", "uint", "uint16", "int32", "uint32", "int64", "byte":  // TODO: Does Joker always have 64-bit signed ints?
			jok = "Int"
			gol = "int"
			return
		case "bool":
			jok = "Bool"
			gol = "bool"
			return
		case "error":
			jok = "Error"
			gol = "error"
			return
		default:
			jok, _ = genGoPostNamed(indent, pkg, tmp, v.Name)
			gol = v.Name  // This is as far as Go needs to go for a type signature
			return
		}
	case *ArrayType:
		jok, gol = genGoPostElement(indent, pkg, tmp, v.Elt)
		jok = "(vector-of " + jok + ")"
		gol = "[]" + gol
		return
	case *StarExpr:
		jok, gol = genGoPostElement(indent, pkg, tmp, v.X)  // TODO: Maybe return a ref or something Joker (someday) supports?
		gol = "*" + gol
	case *StructType:
		jok, gol = genGoPostStruct(indent, pkg, tmp, v.Fields)
	default:
		jok = fmt.Sprintf("ABEND883(unrecognized Expr type %T at: %s)", e, whereAt(e.Pos()))
		return
	}
	return
}

func genGoPostItem(indent, pkg string, f *Field, idx *int, p *Ident, gores, jok, gol, goc *string, multiple *bool) {
	var rtn string
	*idx += 1
	if (*idx > 1) {
		*gores += ", "
	}
	if p == nil || p.Name == "" {
		rtn = fmt.Sprintf("arg_%d", *idx)
	} else {
		rtn = p.Name
	}
	*gores += rtn
	joktype, goltype := genGoPostElement(indent, pkg, rtn, f.Type) // ~~~
	if *jok != "" {
		*jok += " "
		*multiple = true
	}
	*jok += joktype  // TODO: Someday '"^" + joktype + " " + rtn', but only after generate-std.joke supports it
	if *gol != "" {
		*gol += ", "
	}
	if p == nil {
	} else {
		if false {  // TODO: Someday enable this code, but only after generate-std.joke supports it
			if joktype != "" {
				*jok += " "
			}
			*jok += paramNameAsClojure(p.Name)
		}
		*gol += paramNameAsGo(p.Name)
		if goltype != "" {
			*gol += " "
		}
	}
	*gol += goltype
}

func genGoPostList(indent string, pkg string, fl *FieldList) (gores, jok, gol, goc string) {
	multiple := false
	idx := 0
	for _, f := range fl.List {
		if f.Names == nil {
			genGoPostItem(indent, pkg, f, &idx, nil, &gores, &jok, &gol, &goc, &multiple)
		}
		for _, p := range f.Names {
			genGoPostItem(indent, pkg, f, &idx, p, &gores, &jok, &gol, &goc, &multiple)
		}
	}
	if multiple {
		jok = "[" + jok + "]"
		if gol != "" {
			gol = "(" + gol + ")"
		}
	}
	if goc != "" {
		goc = indent + goc
	}
	return
}

// Return a form of the return type as supported by generate-std.joke,
// or empty string if not supported (which will trigger attempting to
// generate appropriate code for *_native.go). gol either passes
// through or "Object" is returned for it if jok is returned as empty.
func jokerReturnTypeForGenerateSTD(in_jok, in_gol string) (jok, gol string) {
	switch in_jok {
	case "String", "Int", "Double", "Bool", "Time", "Error":  // TODO: Have tested only String so far
		jok = "^" + in_jok
	default:
		jok = ""
		gol = "Object"
	}
	return
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

type funcCode struct {
	jokerParamList string  // fieldListAsClojure(d.Type.Params)
	goParamList string  // paramListAsGo(d.Type.Params)
	jokerGoCode string  // goFname + "(" + fieldListToGo(d.Type.Params) + ")"
	goCode string
	jokerReturnTypeForDoc string  // genReturnType(pkg, d.Type.Results)
	goReturnTypeForDoc string  // genReturnType(pkg, d.Type.Results)
}

func genGoPre(indent string, fl *FieldList, goFname string) (jok, jok2gol, gol, code, params string) {
	jok = fieldListAsClojure(fl)
	jok2gol = goFname + "(" + fieldListToGo(fl) + ")"
	code = "" // TODO: enhance to support composites
	gol = paramListAsGo(fl)
	params = argsAsGo(fl)
	return
}

func genGoCall(pkg, goFname string, goParams string) string {
	return pkg + "." + goFname + "(" + goParams + ")\n"
}

func genGoPost(indent string, pkg string, d *FuncDecl) (goResultAssign, jokerReturnTypeForDoc, goReturnTypeForDoc string, goReturnCode string) {
	fl := d.Type.Results
	if fl == nil || fl.List == nil {
		return
	}
	goResultAssign, jokerReturnTypeForDoc, goReturnTypeForDoc, goReturnCode = genGoPostList(indent, pkg, fl)
	return
}

func genFuncCode(pkg string, d *FuncDecl, goFname string) (fc funcCode) {
	var goPreCode, goParams, goResultAssign, goPostCode string

	fc.jokerParamList, fc.jokerGoCode, fc.goParamList, goPreCode, goParams =
		genGoPre("\t", d.Type.Params, goFname)
	goCall := genGoCall(pkg, d.Name.Name, goParams)
	goResultAssign, fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc, goPostCode = genGoPost("\t", pkg, d)

	if goPostCode == "" {
		goPostCode = "\t...ABEND676: TODO..."
	}

	if goResultAssign != "" {
		goResultAssign += " := "
	}
	fc.goCode = goPreCode + // Optional block of pre-code
		"\t" + goResultAssign + goCall + // [results := ]fn-to-call([args...])
		goPostCode // Optional block of post-code
	return
}

func genFunction(f string, fn *funcInfo) {
	d := fn.fd
	pkg := filepath.Base(fn.pkg)
	jfmt := `
(defn %s%s
%s  {:added "1.0"
   :go "%s"}
  [%s])
`
	goFname := funcNameAsGoPrivate(d.Name.Name)
	fc := genFuncCode(pkg, d, goFname)
	jokerReturnType, goReturnType := jokerReturnTypeForGenerateSTD(fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc)
	if jokerReturnType != "" {
		jokerReturnType += " "
	}
	jokerFn := fmt.Sprintf(jfmt, jokerReturnType, d.Name.Name,
		commentGroupInQuotes(d.Doc, fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc),
		fc.jokerGoCode, fc.jokerParamList)

	gfmt := `
func %s(%s) %s {
%s
}
`

	goFn := ""
	if jokerReturnType == "" {  // TODO: Generate this anyway if it contains ABEND, so we can see what's needed.
		goFn = fmt.Sprintf(gfmt, goFname, fc.goParamList, goReturnType, fc.goCode)
	}

	if strings.Contains(jokerFn, "ABEND") || strings.Contains(goFn, "ABEND") {
		jokerFn = nonEmptyLineRegexp.ReplaceAllString(jokerFn, `;; $1`)
		goFn = nonEmptyLineRegexp.ReplaceAllString(goFn, `// $1`)
	}

	if _, ok := jokerCode[pkg]; !ok {
		jokerCode[pkg] = codeInfo {}
	}
	jokerCode[pkg][d.Name.Name] = jokerFn

	if goFn != "" {
		if _, ok := goCode[pkg]; !ok {
			goCode[pkg] = codeInfo {}
		}
		goCode[pkg][d.Name.Name] = goFn
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

	/* Generate function code snippets in alphabetical order, to stabilize test output in re unsupported types. */
	sortedFuncInfoMap(functions,
		func(f string, v *funcInfo) {
			if v.fd != DUPLICATEFUNCTION {
				genFunction(f, v)
			}
		})

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
	. "github.com/candid82/joker/core"
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
