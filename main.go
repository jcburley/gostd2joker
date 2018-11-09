package main

import (
	"bufio"
	"fmt"
	. "go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"regexp"
	"strings"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"time"
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

/* Maintain a set of packages seen, keyed by (relative) package pathname. */
var exists = struct{}{}
var packagesSet = map[string]struct{} {}

/* Sort the packages -- currently appears to not actually be
/* necessary, probably because of how walkDirs() works. */
func sortedPackages(m map[string]struct{}, f func(k string)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k)
	}
}

/* Maps simple package names to their (relative) source directories. */
var packageDirs = map[string]string {}

func processPackage(pkgDir string, pkg string, p *Package) {
	if verbose {
		fmt.Printf("Processing package=%s in %s:\n", pkg, pkgDir)
	}
	if pd, ok := packageDirs[pkg]; ok {
		fmt.Fprintf(os.Stderr,
			"Skipping %s as it was already processed in %s before being seen in %s.\n",
			pkg, pd, pkgDir)
		return
	}
	packageDirs[pkg] = pkgDir
	for filename, f := range p.Files {
		processDecls(pkg, filename, f)
	}
}

func processDir(d string, path string, mode parser.Mode) error {
	pkgDir := strings.TrimPrefix(path, d + string(filepath.Separator))
	if verbose {
		fmt.Printf("Processing %s:\n", pkgDir)
	}
	packagesSet[pkgDir] = exists

	pkgs, err := parser.ParseDir(fset, path,
		// Walk only *.go files that meet default (target) build constraints, e.g. per "// build ..."
		func (info os.FileInfo) bool {
			if strings.HasSuffix(info.Name(), "_test.go") {
				if verbose {
					fmt.Printf("Ignoring test code in %s\n", info.Name())
				}
				return false
			}
			b, e := build.Default.MatchFile(path, info.Name());
			if verbose {
				fmt.Printf("Matchfile(%s) => %v %v\n", filepath.Join(path, info.Name()), b, e)
			}
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
			processPackage(pkgDir, k, v) // processPackage(strings.Replace(path, d + "/", "", 1) + "/" + k, v)
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
		case "int", "uint", "int16", "uint16", "int32", "uint32", "int64":
			return "Int"
		case "byte":
			return "Byte"
		case "bool":
			return "Bool"
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
		case "string", "int", "int16", "uint", "uint16", "int32", "uint32", "int64", "byte", "bool", "error":
			return v.Name
		default:
			return fmt.Sprintf("ABEND884(unrecognized type %s at: %s)", v.Name, whereAt(e.Pos()))
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

var genSymIndex = 0

func genSym(pre string) string {
	genSymIndex += 1
	return fmt.Sprintf("%s%d", pre, genSymIndex)
}

func genSymReset() {
	genSymIndex = 0
}

func genGoPostNamed(indent, pkg, in, t string) (jok, gol, goc, out string) {
	qt := pkg + "." + t
	if v, ok := types[qt]; ok {
		if v[0].building { // Mutually-referring types currently not supported
			jok = fmt.Sprintf("ABEND947(recursive type reference involving %s)", qt)  // TODO: handle these, e.g. http Request/Response
			gol = jok
			goc = ""
		} else {
			v[0].building = true
			jok, gol, goc, out = genGoPostExpr(indent, pkg, in, v[0].td.Type)
			v[0].jok = jok
			v[0].gol = gol
			v[0].building = false
		}
	} else {
		jok = fmt.Sprintf("ABEND042(cannot find typename %s)", qt)
	}
	return
}

// func tryThis(s string) struct { a int; b string } {
//	return struct { a int; b string }{ 5, "hey" }
// }

// Joker: { :a ^Int, :b ^String }
// Go: struct { a int; b string }
func genGoPostStruct(indent, pkg, in string, fl *FieldList) (jok, gol, goc, out string) {
	if fl == nil {
		jok = "{}"
		gol = "struct {}"
		out = in
		return
	}
	tmpmap := genSym("map")
	goc += indent + tmpmap + " := EmptyArrayMap()\n"
	for _, f := range fl.List {
		for _, p := range f.Names {
			var joktype, goltype, more_goc string
			joktype, goltype, more_goc, out =
				genGoPostExpr(indent, pkg, in + "." + p.Name, f.Type)
			goc += more_goc
			goc += indent + tmpmap + ".Add(MakeKeyword(\"" + p.Name + "\"), " + out + ")\n"
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
	out = tmpmap
	return
}

func genGoPostArray(indent, pkg, in string, el Expr) (jok, gol, goc, out string) {
	tmp := genSym("")
	tmpvec := "vec" + tmp
	tmpelem := "elem" + tmp
	goc += indent + tmpvec + " := EmptyVector\n"
	goc += indent + "for _, " + tmpelem + " := range " + in + " {\n"

	var goc_pre string
	jok, gol, goc_pre, out = genGoPostExpr(indent + "\t", pkg, tmpelem, el)
	jok = "(vector-of " + jok + ")"
	gol = "[]" + gol

	goc += goc_pre
	goc += indent + "\t" + tmpvec + " = " + tmpvec + ".Conjoin(" + out + ")\n"
	goc += indent + "}\n"
	out = tmpvec
	return
}

// TODO: Maybe return a ref or something Joker (someday) supports?
func genGoPostStar(indent, pkg, in string, e Expr) (jok, gol, goc, out string) {
	var new_out string
	jok, gol, goc, new_out = genGoPostExpr(indent, pkg, "(*" + in + ")", e)
	out = new_out
	gol = "*" + gol
	return
}

func maybeNil(expr, in string) string {
	return "func () Object { if (" + expr + ") == nil { return NIL } else { return " + in + " } }()"
}

func genGoPostExpr(indent, pkg, in string, e Expr) (jok, gol, goc, out string) {
	switch v := e.(type) {
	case *Ident:
		switch v.Name {
		case "string":
			jok = "String"
			gol = "string"
			out = "MakeString(" + in + ")"
		case "int", "int16", "uint", "uint16", "int32", "uint32", "int64", "byte":  // TODO: Does Joker always have 64-bit signed ints?
			jok = "Int"
			gol = "int"
			out = "MakeInt(int(" + in + "))"
		case "bool":
			jok = "Bool"
			gol = "bool"
			out = "MakeBool(" + in + ")"
		case "error":
			jok = "Error"
			gol = "error"
			out = maybeNil(in, "MakeString(" + in + ".Error())")  // TODO: Test this, as I can't find a MakeError() in joker/core/object.go
		default:
			jok, _, goc, out = genGoPostNamed(indent, pkg, in, v.Name)
			gol = v.Name  // This is as far as Go needs to go for a type signature
		}
	case *ArrayType:
		jok, gol, goc, out = genGoPostArray(indent, pkg, in, v.Elt)
	case *StarExpr:
		jok, gol, goc, out = genGoPostStar(indent, pkg, in, v.X)
	case *StructType:
		jok, gol, goc, out = genGoPostStruct(indent, pkg, in, v.Fields)
	default:
		jok = fmt.Sprintf("ABEND883(unrecognized Expr type %T at: %s)", e, whereAt(e.Pos()))
		gol = "..."
		out = in
	}
	return
}

func genGoPostItem(indent, pkg string, f *Field, idx *int, p *Ident, gores, jok, gol, goc *string) string {
	var rtn string
	*idx += 1
	if (*idx > 1) {
		*gores += ", "
	}
	if gores == nil {
		rtn = ""
	} else if *gores == "res" {
		rtn = *gores
	} else {
		if p == nil || p.Name == "" {
			rtn = fmt.Sprintf("res%d", *idx)
		} else {
			rtn = p.Name
		}
		*gores += rtn
	}
	joktype, goltype, goc_pre, val := genGoPostExpr(indent, pkg, rtn, f.Type)
	*goc = goc_pre + *goc
	if *jok != "" {
		*jok += " "
	}
	*jok += joktype  // TODO: Someday '"^" + joktype', but only after generate-std.joke supports it
	if *gol != "" {
		*gol += ", "
	}
	if p != nil {
		if false {  // TODO: Someday enable this code, but only after generate-std.joke supports it
			*jok += " " + paramNameAsClojure(p.Name)
		}
		*gol += paramNameAsGo(p.Name) + " "
	}
	*gol += goltype
	return val
}

func genGoPostList(indent string, pkg string, fl *FieldList) (gores, jok, gol, goc string) {
	idx := 0
	multiple := len(fl.List) > 1 || (fl.List[0].Names != nil && len(fl.List[0].Names) > 1)
	if multiple {
		for _, f := range fl.List {
			if f.Names == nil {
				rtn := genGoPostItem(indent, pkg, f, &idx, nil, &gores, &jok, &gol, &goc)
				goc += indent + "res = res.Conjoin(" + rtn + ")\n"
			} else {
				for _, p := range f.Names {
					rtn := genGoPostItem(indent, pkg, f, &idx, p, &gores, &jok, &gol, &goc)
					goc += indent + "res = res.Conjoin(" + rtn + ")\n"
				}
			}
		}
		jok = "[" + jok + "]"
		if gol != "" {
			gol = "(" + gol + ")"
		}
		goc = indent + "res := EmptyVector\n" + goc + indent + "return res\n"
		gores += " := "
	} else if len(fl.List) == 0 {
		genGoPostItem(indent, pkg, fl.List[0], &idx, nil,
			nil, &jok, &gol, &goc)
	} else {
		gores = "res"
		rtn := genGoPostItem(indent, pkg, fl.List[0], &idx, nil,
			&gores, &jok, &gol, &goc)
		if goc == "" {
			gores = "return " // No code generated, so no need to use rtn as intermediary
		} else {
			goc += indent + "return " + rtn + "\n"
			gores += " := "
		}
	}
	return
}

// Return a form of the return type as supported by generate-std.joke,
// or empty string if not supported (which will trigger attempting to
// generate appropriate code for *_native.go). gol either passes
// through or "Object" is returned for it if jok is returned as empty.
func jokerReturnTypeForGenerateSTD(in_jok, in_gol string) (jok, gol string) {
	switch in_jok {
	case "String", "Int", "Byte", "Double", "Bool", "Time", "Error":  // TODO: Have tested only String so far
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
	jokerGoParams string  // "(" + fieldListToGo(d.Type.Params) + ")"
	goCode string
	jokerReturnTypeForDoc string  // genReturnType(pkg, d.Type.Results)
	goReturnTypeForDoc string  // genReturnType(pkg, d.Type.Results)
}

func genGoPre(indent string, fl *FieldList, goFname string) (jok, jok2golParams, gol, code, params string) {
	jok = fieldListAsClojure(fl)
	jok2golParams = "(" + fieldListToGo(fl) + ")"
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

	fc.jokerParamList, fc.jokerGoParams, fc.goParamList, goPreCode, goParams =
		genGoPre("\t", d.Type.Params, goFname)
	goCall := genGoCall(pkg, d.Name.Name, goParams)
	goResultAssign, fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc, goPostCode = genGoPost("\t", pkg, d)

	if goPostCode == "" && goResultAssign == "" {
		goPostCode = "\t...ABEND675: TODO...\n"
	}

	fc.goCode = goPreCode + // Optional block of pre-code
		"\t" + goResultAssign + goCall + // [results := ]fn-to-call([args...])
		goPostCode // Optional block of post-code
	return
}

func genFunction(f string, fn *funcInfo) {
	genSymReset()
	d := fn.fd
	pkg := filepath.Base(fn.pkg)
	jfmt := `
(defn %s%s
%s  {:added "1.0"
   :go "%s%s"}
  [%s])
`
	goFname := funcNameAsGoPrivate(d.Name.Name)
	fc := genFuncCode(pkg, d, goFname)
	jokerReturnType, goReturnType := jokerReturnTypeForGenerateSTD(fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc)

	var jok2gol string
	if jokerReturnType == "" {
		jok2gol = goFname
	} else {
		jokerReturnType += " "
		jok2gol = pkg + "." + d.Name.Name
	}

	jokerFn := fmt.Sprintf(jfmt, jokerReturnType, d.Name.Name,
		commentGroupInQuotes(d.Doc, fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc),
		jok2gol, fc.jokerGoParams, fc.jokerParamList)

	gfmt := `
func %s(%s) %s {
%s}
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
  --go <go-source-dir-name>      # Location of Go source tree's src/ subdirectory
  --overwrite                    # Overwrite any existing <joker-std-subdir> files, leaving existing files intact
  --replace                      # 'rm -fr <joker-std-subdir>' before creating <joker-std-subdir>
  --fresh                        # (Default) Refuse to overwrite existing <joker-std-subdir> directory
  --joker <joker-source-dir-name>  # Modify pertinent source files to reflect packages being created
  --verbose, -v                  # Print info on what's going on
  --dump                         # Use go's AST dump API on pertinent elements (functions, types, etc.)
  --help, -h                     # Print this information

If <joker-std-subdir> is not specified, no Go nor Clojure source files
(nor any other files nor directories) are created, effecting a sort of
"dry run".
`)
	os.Exit(0)
}

var currentTime = ""
func curTime() string {
	if currentTime == "" {
		by, _ := time.Now().MarshalText()
		currentTime = string(by)
	}
	return currentTime
}

// E.g.: \t_ "github.com/candid82/joker/std/go/net"
func updateJokerMain(pkgs []string, f string) {
	by, err := ioutil.ReadFile(f)
	check(err)
	m := string(by)
	flag := "Imports added by gostd2joker"
	endflag := "End gostd2joker-added imports"

	if !strings.Contains(m, flag) {
		if verbose {
			fmt.Printf("Adding custom import line to %s\n", f)
		}
		m = strings.Replace(m, "import", "import ( // " + flag + "\n) // " + endflag + "\n\nimport", 1)
		m = "// Auto-modified by gostd2joker at " + curTime() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag)  // [^(]*[(][^)]*[)]
	newImports := "\n"
	importPrefix := "\t_ \"github.com/candid82/joker/std/go/"
	for _, p := range pkgs {
		newImports += importPrefix + p + "\"\n"
	}
	m = reImport.ReplaceAllString(m, flag + newImports + ") // " + endflag)

	if verbose {
		fmt.Printf("Writing %s\n", f)
	}
	err = ioutil.WriteFile(f, []byte(m), 0777)
	check(err)
}

// E.g.: *loaded-libs* #{'joker.core 'joker.os 'joker.base64 'joker.json 'joker.string 'joker.yaml 'joker.go.net})
func updateCoreDotJoke(pkgs []string, f string) {
	by, err := ioutil.ReadFile(f)
	check(err)
	m := string(by)
	flag := "Loaded-libraries added by gostd2joker"
	endflag := "End gostd2joker-added loaded-libraries"

	if !strings.Contains(m, flag) {
		if verbose {
			fmt.Printf("Adding custom loaded libraries to %s\n", f)
		}
		m = strings.Replace(m, "\n  *loaded-libs* #{",
			"\n  *loaded-libs* #{\n   ;; " + flag + "\n   ;; " + endflag + "\n", 1)
		m = ";;;; Auto-modified by gostd2joker at " + curTime() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag + "\n *?")
	newImports := "\n  "
	importPrefix := " 'joker.go."
	curLine := ""
	for _, p := range pkgs {
		more := importPrefix + strings.Replace(p, string(filepath.Separator), ".", -1)
		if curLine != "" && len(curLine) + len(more) > 77 {
			newImports += curLine + "\n  "
			curLine = more
		} else {
			curLine += more
		}
	}
	newImports += curLine
	m = reImport.ReplaceAllString(m, flag + newImports + "\n   ;; " + endflag + "\n   ")

	if verbose {
		fmt.Printf("Writing %s\n", f)
	}
	err = ioutil.WriteFile(f, []byte(m), 0777)
	check(err)
}

// E.g.: (def namespaces ['string 'json 'base64 'os 'time 'yaml 'http 'math 'html 'url])
func updateGenerateSTD(pkgs []string, f string) {
	by, err := ioutil.ReadFile(f)
	check(err)
	m := string(by)
	flag := "Namespaces added by gostd2joker"
	endflag := "End gostd2joker-added namespaces"

	if !strings.Contains(m, flag) {
		if verbose {
			fmt.Printf("Adding custom namespaces to %s\n", f)
		}
		m = strings.Replace(m, "(def namespaces [",
			"(def namespaces [\n  ;; " + flag + "\n  ;; " + endflag + "\n  ", 1)
		m = ";;;; Auto-modified by gostd2joker at " + curTime() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag + "\n *?")
	newImports := "\n "
	importPrefix := " 'go."
	curLine := ""
	for _, p := range pkgs {
		more := importPrefix + strings.Replace(p, string(filepath.Separator), ".", -1)
		if curLine != "" && len(curLine) + len(more) > 77 {
			newImports += curLine + "\n "
			curLine = more
		} else {
			curLine += more
		}
	}
	newImports += curLine
	m = reImport.ReplaceAllString(m, flag + newImports + "\n  ;; " + endflag + "\n  ")

	if verbose {
		fmt.Printf("Writing %s\n", f)
	}
	err = ioutil.WriteFile(f, []byte(m), 0777)
	check(err)
}

func main() {
	fset = token.NewFileSet() // positions are relative to fset
	dump = false

	length := len(os.Args)
	sourceDir := ""
	jokerSourceDir := ""
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
			case "--go":
				if sourceDir != "" {
					panic("cannot specify --go <go-source-dir-name> more than once")
				}
				if i < length-1 && notOption(os.Args[i+1]) {
					i += 1 // shift
					sourceDir = os.Args[i]
				} else {
					panic("missing path after --go option")
				}
			case "--joker":
				if jokerSourceDir != "" {
					panic("cannot specify --joker <joker-source-dir-name> more than once")
				}
				if i < length-1 && notOption(os.Args[i+1]) {
					i += 1 // shift
					jokerSourceDir = os.Args[i]
				} else {
					panic("missing path after --joker option")
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
		goLink := "GO.link"
		if si, e := os.Stat(goLink); e != nil || !si.IsDir() {
			panic("Must specify --go <go-source-dir-name> option, or make ./GO.link a symlink to the golang/go/ source directory")
		}
	}

	sourceDir = filepath.Join(sourceDir, "src")

	if fi, e := os.Stat(filepath.Join(sourceDir, "go")); e != nil || !fi.IsDir() {
		if m, e := filepath.Glob(filepath.Join(sourceDir, "*.go")); e != nil || m == nil || len(m) == 0 {
			panic(fmt.Sprintf("Does not exist or is not a Go source directory: %s;\n%v", sourceDir, m))
		}
	}

	jokerLibDir := ""
	if jokerSourceDir != "" && jokerSourceDir != "-" {
		jokerLibDir = filepath.Join(jokerSourceDir, "std", "go")
		if replace {
			if e := os.RemoveAll(jokerLibDir); e != nil {
				panic(fmt.Sprintf("Unable to effectively 'rm -fr %s'", jokerLibDir))
			}
		}

		if !overwrite {
			var stat syscall.Stat_t
			if e := syscall.Stat(jokerLibDir, &stat); e == nil || e.Error() != "no such file or directory" {
				msg := "already exists"
				if e != nil {
					msg = e.Error()
				}
				panic(fmt.Sprintf("Refusing to populate existing directory %s; please 'rm -fr' first, or specify --overwrite or --replace: %s",
					jokerLibDir, msg))
			}
			if e := os.MkdirAll(jokerLibDir, 0777); e != nil {
				panic(fmt.Sprintf("Cannot 'mkdir -p %s': %s", jokerLibDir, e.Error()))
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
			if jokerLibDir != "" && jokerLibDir != "-" {
				jf := filepath.Join(jokerLibDir, p + ".joke")
				var e error
				unbuf_out, e = os.Create(jf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)
				fmt.Fprintf(out, `;;;; Auto-generated by gostd2joker at ` + curTime() + `, do not edit!!

(ns
  ^{:go-imports ["%s"]
    :doc "Provides a low-level interface to the %s package."}
  go.%s)
`,
					p, p, p)
			}
			sortedCodeMap(v,
				func(f string, w string) {
					if verbose || jokerLibDir == "" {
						fmt.Printf("JOKER FUNC %s.%s has:%v\n", p, f, w)
					}
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
			if jokerLibDir != "" && jokerLibDir != "-" {
				gf := filepath.Join(jokerLibDir, packageDirs[p], p + "_native.go")
				var e error
				e = os.MkdirAll(filepath.Dir(gf), 0777)
				check(e)
				unbuf_out, e = os.Create(gf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)
				fmt.Fprintf(out, `// Auto-generated by gostd2joker at ` + curTime() + `, do not edit!!

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
					if verbose || jokerLibDir == "" {
						fmt.Printf("GO FUNC %s.%s has:%v\n", p, f, w)
					}
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

	if jokerSourceDir != "" && jokerSourceDir != "-" {
		var packagesArray = []string{} // Relative package pathnames in alphabetical order

		sortedPackages(packagesSet,
			func (p string) { packagesArray = append(packagesArray, p) })
		updateJokerMain(packagesArray, filepath.Join(jokerSourceDir, "main.go"))
		updateCoreDotJoke(packagesArray, filepath.Join(jokerSourceDir, "core", "data", "core.joke"))
		updateGenerateSTD(packagesArray, filepath.Join(jokerSourceDir, "std", "generate-std.joke"))
	}

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
