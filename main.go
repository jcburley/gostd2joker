package main

import (
	"bufio"
	"fmt"
	. "go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

func unix(p string) string {
	return filepath.ToSlash(p)
}

func commentGroupInQuotes(doc *CommentGroup, jok, gol string) string {
	var d string
	if doc != nil {
		d = doc.Text()
	}
	if gol != "" {
		if d != "" {
			d = strings.Trim(d, " \t\n") + "\n\n"
		}
		d += "Go return type: " + gol
	}
	if jok != "" {
		if d != "" {
			d = strings.Trim(d, " \t\n") + "\n\n"
		}
		d += "Joker return type: " + jok
	}
	return `  ` + strings.Trim(strconv.Quote(d), " \t\n") + "\n"
}

type funcInfo struct {
	fd *FuncDecl
	pkg string // base package name
	pkgDirUnix string // relative (Unix-style) path to package
	filename string // relative (Unix-style) filename within package
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

// Map (unqualified) function names to info on each.
var unqualifiedFunctions = map[string]*funcInfo {}
var DUPLICATEFUNCTION = &FuncDecl {}

var alreadySeen = []string {}

// Returns whether any public functions were actually processed.
func processFuncDecl(pkg, pkgDirUnix, filename string, f *File, fn *FuncDecl) bool {
	if (dump) {
		Print(fset, fn)
	}
	fname := pkgDirUnix + "." + fn.Name.Name
	if v, ok := unqualifiedFunctions[fname]; ok {
		if v.fd != DUPLICATEFUNCTION {
			alreadySeen = append(alreadySeen,
				fmt.Sprintf("NOTE: Already seen function %s in %s, yet again in %s",
					fname, v.filename, filename))
			fn = DUPLICATEFUNCTION
		}
	}
	unqualifiedFunctions[fname] = &funcInfo{fn, pkg, pkgDirUnix, filename}
	return true
}

type typeInfo struct {
	td *TypeSpec
	file string // Relative (Unix-style) path to defining file
	building bool
}

func sortedTypeInfoMap(m map[string]*typeInfo, f func(k string, v *typeInfo)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}

var types = map[string]*typeInfo {}

func processTypeSpec(pkg string, filename string, f *File, ts *TypeSpec) {
	if (dump) {
		Print(fset, ts)
	}
	typename := pkg + "." + ts.Name.Name
	if c, ok := types[typename]; ok {
		if c.file == filename {
			panic(fmt.Sprintf("type %s defined twice in file %s", typename, filename))
		}
	}
	types[typename] = &typeInfo{ts, filename, false}
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

// Returns whether any public functions were actually processed.
func processDecls(pkg, pkgDirUnix, filename string, f *File) (found bool) {
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
			if processFuncDecl(pkg, pkgDirUnix, filename, f, v) {
				found = true
			}
		case *GenDecl:
			if v.Tok != token.TYPE {
				continue
			}
			processTypeSpecs(pkgDirUnix, filename, f, v.Specs)
		default:
			panic(fmt.Sprintf("unrecognized Decl type %T at: %s", v, whereAt(v.Pos())))
		}
	}
	return
}

var exists = struct{}{}

/* Maps relative package (unix-style) names to their imports, non-emptiness, etc. */
type packageImports map[string]struct{}
type packageInfo struct {
	importsNative packageImports
	importsAutoGen packageImports
	nonEmpty bool // Whether any non-comment code has been generated
	hasGoFiles bool // Whether any .go files (would) have been generated
}
var packagesInfo = map[string]*packageInfo {}

/* Sort the packages -- currently appears to not actually be
/* necessary, probably because of how walkDirs() works. */
func sortedPackagesInfo(m map[string]*packageInfo, f func(k string, i *packageInfo)) {
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k, m[k])
	}
}

func sortedPackageImports(pi packageImports, f func(k string)) {
	var keys []string
	for k, _ := range pi {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f(k)
	}
}

func processPackage(pkgDir, pkgDirUnix, pkg string, p *Package) {
	if verbose {
		fmt.Printf("Processing package=%s in %s:\n", pkg, pkgDirUnix)
	}
	found := false
	for filename, f := range p.Files {
		if processDecls(pkg, pkgDirUnix, filepath.ToSlash(filename), f) {
			found = true
		}
	}
	if found {
		if _, ok := packagesInfo[pkgDirUnix]; !ok {
			packagesInfo[pkgDirUnix] = &packageInfo{packageImports{}, packageImports{}, false, false}
		}
	}
}

func processDir(d string, path string, mode parser.Mode) error {
	pkgDir := strings.TrimPrefix(path, d + string(filepath.Separator))
	pkgDirUnix := filepath.ToSlash(pkgDir)
	if verbose {
		fmt.Printf("Processing %s:\n", pkgDirUnix)
	}

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
				fmt.Printf("Matchfile(%s) => %v %v\n",
					filepath.ToSlash(filepath.Join(path, info.Name())),
					b, e)
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
			processPackage(pkgDir, pkgDirUnix, k, v) // processPackage(strings.Replace(path, d + "/", "", 1) + "/" + k, v)
		}
	}

	return nil
}

var excludeDirs = map[string]bool {
	"builtin": true,
	"cmd": true,
	"internal": true, // look into this later?
	"testdata": true,
	"vendor": true,
}


func walkDirs(d string, mode parser.Mode) error {
	target, err := filepath.EvalSymlinks(d)
	check(err)
	err = filepath.Walk(target,
		func(path string, info os.FileInfo, err error) error {
			rel := strings.Replace(path, target, d, 1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s due to: %v\n",
					filepath.ToSlash(rel), err)
				return err
			}
			if rel == d {
				return nil // skip (implicit) "."
			}
			if excludeDirs[filepath.Base(rel)] {
				if verbose {
					fmt.Printf("Excluding %s\n",
						filepath.ToSlash(rel))
				}
				return filepath.SkipDir
			}
			if info.IsDir() {
				if verbose {
					fmt.Printf("Walking from %s to %s\n",
						filepath.ToSlash(d), filepath.ToSlash(rel))
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
		case "int":
			return "Int"
		case "byte":
			return "Byte"
		case "bool":
			return "Bool"
		default:
			return fmt.Sprintf("ABEND885(unrecognized type %s at: %s)", v.Name, whereAt(e.Pos()))
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
				s += "_" + paramNameAsClojure(p.Name)
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
				s += "_" + p.Name
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

var genSymIndex = map[string]int {}

func genSym(pre string) string {
	var idx int
	if i, ok := genSymIndex[pre]; ok {
		idx = i + 1
	} else {
		idx = 1
	}
	genSymIndex[pre] = idx
	return fmt.Sprintf("%s%d", pre, idx)
}

func genSymReset() {
	genSymIndex = map[string]int {}
}

func exprIsUseful(rtn string) bool {
	return rtn != "NIL"
}

func genGoPostNamed(indent, pkg, in, t, onlyIf string) (jok, gol, goc, out string) {
	qt := pkg + "." + t
	if v, ok := types[qt]; ok {
		if v.building { // Mutually-referring types currently not supported
			jok = fmt.Sprintf("ABEND947(recursive type reference involving %s)",
				qt)  // TODO: handle these, e.g. http Request/Response
			gol = jok
			goc = ""
		} else {
			v.building = true
			jok, gol, goc, out = genGoPostExpr(indent, pkg, in, v.td.Type, onlyIf)
			v.building = false
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
func genGoPostStruct(indent, pkg, in string, fl *FieldList, onlyIf string) (jok, gol, goc, out string) {
	tmpmap := "_map" + genSym("")
	useful := false
	for _, f := range fl.List {
		for _, p := range f.Names {
			if unicode.IsLower(rune(p.Name[0])) {
				continue  // Skipping non-exported fields
			}
			var joktype, goltype, more_goc string
			joktype, goltype, more_goc, out =
				genGoPostExpr(indent, pkg, in + "." + p.Name, f.Type, "")
			if useful || exprIsUseful(out) {
				useful = true
			}
			goc += more_goc
			goc += indent + tmpmap +
				".Add(MakeKeyword(\"" + p.Name + "\"), " + out + ")\n"
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
	if useful {
		goc = wrapStmtOnlyIfs(indent, tmpmap, "ArrayMap", "EmptyArrayMap()", onlyIf, goc, &out)
	} else {
		goc = ""
		out = "NIL"
	}
	return
}

func genGoPostArray(indent, pkg, in string, el Expr, onlyIf string) (jok, gol, goc, out string) {
	tmp := genSym("")
	tmpvec := "_vec" + tmp
	tmpelem := "_elem" + tmp

	var goc_pre string
	jok, gol, goc_pre, out = genGoPostExpr(indent + "\t", pkg, tmpelem, el, "")
	useful := exprIsUseful(out)
	jok = "(vector-of " + jok + ")"
	gol = "[]" + gol

	if useful {
		goc = indent + "for _, " + tmpelem + " := range " + in + " {\n"
		goc += goc_pre
		goc += indent + "\t" + tmpvec + " = " + tmpvec + ".Conjoin(" + out + ")\n"
		goc += indent + "}\n"
		goc = wrapStmtOnlyIfs(indent, tmpvec, "Vector", "EmptyVector", onlyIf, goc, &out)
	} else {
		goc = ""
	}
	return
}

// TODO: Maybe return a ref or something Joker (someday) supports? flag.String() is useful only as it returns a ref;
// whereas net.LookupMX() returns []*MX, and these are not only populated, it's unclear there's any utility in
// modifying them (it could just as well return []MX AFAICT).
func genGoPostStar(indent, pkg, in string, e Expr, onlyIf string) (jok, gol, goc, out string) {
	if onlyIf != "" {
		panic(fmt.Sprintf("pkg=%s in=%s e=%v onlyIf=%s\n", pkg, in, e, onlyIf))
	}
	jok, gol, goc, out = genGoPostExpr(indent, pkg, "(*" + in + ")", e, in + " != nil")
	gol = "*" + gol
	return
}

func maybeNil(expr, in string) string {
	return "func () Object { if (" + expr + ") == nil { return NIL } else { return " + in + " } }()"
}

func genGoPostExpr(indent, pkg, in string, e Expr, onlyIf string) (jok, gol, goc, out string) {
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
			out = maybeNil(in, "MakeError(" + in + ")")  // TODO: Test this against the MakeError() added to joker/core/object.go
		default:
			jok, _, goc, out = genGoPostNamed(indent, pkg, in, v.Name, onlyIf)
			gol = v.Name  // This is as far as Go needs to go for a type signature
		}
	case *ArrayType:
		jok, gol, goc, out = genGoPostArray(indent, pkg, in, v.Elt, onlyIf)
	case *StarExpr:
		jok, gol, goc, out = genGoPostStar(indent, pkg, in, v.X, onlyIf)
	case *StructType:
		jok, gol, goc, out = genGoPostStruct(indent, pkg, in, v.Fields, onlyIf)
	default:
		jok = fmt.Sprintf("ABEND883(unrecognized Expr type %T at: %s)", e, unix(whereAt(e.Pos())))
		gol = "..."
		out = in
	}
	return
}

const resultName = "_res"

func genGoPostItem(indent, pkg, in string, f *Field, onlyIf string) (captureVar, jok, gol, goc, out string, useful bool) {
	captureVar = in
	if in == "" {
		captureVar = genSym(resultName)
	}
	jok, gol, goc, out = genGoPostExpr(indent, pkg, captureVar, f.Type, onlyIf)
	if in != "" && in != resultName {
		gol = paramNameAsGo(in) + " " + gol
	}
	useful = exprIsUseful(out)
	if !useful {
		captureVar = "_"
	}
	return
}

func reverseJoin(a []string, infix string) string {
	j := ""
	for idx := len(a) - 1; idx >= 0; idx-- {
		if idx != len(a) - 1 {
			j += infix
		}
		j += a[idx]
	}
	return j
}

// Generates code that, at run time, tests each of the onlyIf's and, if all true, returns the expr; else returns NIL.
func wrapOnlyIfs(onlyIf string, e string) string {
	if len(onlyIf) == 0 {
		return e
	}
	return "func() Object { if " + onlyIf + " { return " + e + " } else { return NIL } }()"
}

// Add one level of indent to each line
func indentedCode(c string) string {
	return "\t" + strings.Replace(c, "\n", "\n\t", -1)
}

func wrapStmtOnlyIfs(indent, v, t, e string, onlyIf string, c string, out *string) string {
	if len(onlyIf) == 0 {
		*out = v
		return indent + v + " := " + e + "\n" + c
	}
	*out = "_obj" + v
	return indent + "var " + *out + " Object\n" +
		indent + "if " + onlyIf + " {\n" +
		indent + "\t" + v + " := " + e + "\n" +
		strings.TrimRight(indentedCode(c), "\t") +
		indent + "\t" + *out + " = Object(" + v + ")\n" +
		indent + "} else {\n" +
		indent + "\t" + *out + " = NIL\n" +
		indent + "}\n"
}

// Caller generates "outGOCALL;goc" while saving jok and gol for type info (they go into .joke as metadata and docstrings)
func genGoPostList(indent string, pkg string, fl FieldList) (jok, gol, goc, out string) {
	useful := false
	captureVars := []string {}
	jokType := []string {}
	golType := []string {}
	goCode := []string {}

	result := resultName
	multipleCaptures := len(fl.List) > 1 || (fl.List[0].Names != nil && len(fl.List[0].Names) > 1)
	for _, f := range fl.List {
		names := []string {}
		if f.Names == nil {
			names = append(names, "")
		} else {
			for _, n := range f.Names {
				names = append(names, n.Name)
			}
		}
		for _, n := range names {
			captureName := result
			if multipleCaptures {
				captureName = n
			}
			captureVar, jok, gol, goc, out, usefulItem := genGoPostItem(indent, pkg, captureName, f, "")
			useful = useful || usefulItem
			if multipleCaptures {
				goc += indent + result + " = " + result + ".Conjoin(" + out + ")\n"
			} else {
				result = out
			}
			captureVars = append(captureVars, captureVar)
			jokType = append(jokType, jok)
			golType = append(golType, gol)
			goCode = append(goCode, goc)
		}
	}

	out = strings.Join(captureVars, ", ")
	if out != "" {
		out += " := "
	}

	jok = strings.Join(jokType, " ")
	if len(jokType) > 1 && jok != "" {
		jok = "[" + jok + "]"
	}

	gol = strings.Join(golType, ", ")
	if len(golType) > 1 && gol != "" {
		gol = "(" + gol + ")"
	}

	goc = strings.Join(goCode, "")

	if multipleCaptures {
		if useful {
			goc = indent + result + " := EmptyVector\n" + goc + indent + "return " + result + "\n"
		} else {
			goc = indent + "ABEND123(no public information returned)\n"
		}
	} else {
		if goc == "" {
			out = "return " // No code generated, so no need to use intermediary
		} else {
			goc += indent + "return " + result + "\n"
		}
		if !useful {
			goc += indent + "ABEND124(no public information returned)\n"
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
		jok = `^"` + in_jok + `"`
	default:
		jok = ""
		gol = "Object"
	}
	return
}

type codeInfo map[string]string

/* Map relative (Unix-style) package names to maps of filenames to code strings. */
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
	return "_" + pkg + "." + goFname + "(" + goParams + ")\n"
}

func genGoPost(indent string, pkg string, d *FuncDecl) (goResultAssign, jokerReturnTypeForDoc, goReturnTypeForDoc string, goReturnCode string) {
	fl := d.Type.Results
	if fl == nil || fl.List == nil {
		return
	}
//	goResultAssign, jokerReturnTypeForDoc, goReturnTypeForDoc, goReturnCode = genGoPostListOld(indent, pkg, *fl)
	jokerReturnTypeForDoc, goReturnTypeForDoc, goReturnCode, goResultAssign = genGoPostList(indent, pkg, *fl)
	return
}

func genFuncCode(pkgBaseName, pkgDirUnix string, d *FuncDecl, goFname string) (fc funcCode) {
	var goPreCode, goParams, goResultAssign, goPostCode string

	fc.jokerParamList, fc.jokerGoParams, fc.goParamList, goPreCode, goParams =
		genGoPre("\t", d.Type.Params, goFname)
	goCall := genGoCall(pkgBaseName, d.Name.Name, goParams)
	goResultAssign, fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc, goPostCode =
		genGoPost("\t", pkgDirUnix, d)

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
	pkgDirUnix := fn.pkgDirUnix
	pkgBaseName := filepath.Base(pkgDirUnix)
	jfmt := `
(defn %s%s
%s  {:added "1.0"
   :go "%s%s"}
  [%s])
`
	goFname := funcNameAsGoPrivate(d.Name.Name)
	fc := genFuncCode(pkgBaseName, pkgDirUnix, d, goFname)
	jokerReturnType, goReturnType := jokerReturnTypeForGenerateSTD(fc.jokerReturnTypeForDoc, fc.goReturnTypeForDoc)

	var jok2gol string
	if jokerReturnType == "" {
		jok2gol = goFname
	} else {
		jokerReturnType += " "
		jok2gol = pkgBaseName + "." + d.Name.Name
		if _, found := packagesInfo[pkgDirUnix]; !found {
			panic(fmt.Sprintf("Cannot find package %s", pkgDirUnix))
		}
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
	} else {
		packagesInfo[pkgDirUnix].nonEmpty = true
		if jokerReturnType == "" {
			packagesInfo[pkgDirUnix].importsNative[pkgDirUnix] = exists
		} else {
			packagesInfo[pkgDirUnix].importsAutoGen[pkgDirUnix] = exists
		}
	}

	if _, ok := jokerCode[pkgDirUnix]; !ok {
		jokerCode[pkgDirUnix] = codeInfo {}
	}
	jokerCode[pkgDirUnix][d.Name.Name] = jokerFn

	if goFn != "" {
		if _, ok := goCode[pkgDirUnix]; !ok {
			goCode[pkgDirUnix] = codeInfo {}
		}
		goCode[pkgDirUnix][d.Name.Name] = goFn
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
  --no-timestamp                 # Don't put the time (and version) info in generated/modified files
  --help, -h                     # Print this information

If <joker-std-subdir> is not specified, no Go nor Clojure source files
(nor any other files nor directories) are created, effecting a sort of
"dry run".
`)
	os.Exit(0)
}

var currentTimeAndVersion = ""
var noTimeAndVersion = false
func curTimeAndVersion() string {
	if noTimeAndVersion {
		return "(omitted for testing)"
	}
	if currentTimeAndVersion == "" {
		by, _ := time.Now().MarshalText()
		currentTimeAndVersion = string(by) + " by version " + VERSION
	}
	return currentTimeAndVersion
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
			fmt.Printf("Adding custom import line to %s\n", filepath.ToSlash(f))
		}
		m = strings.Replace(m, "import", "import ( // " + flag + "\n) // " + endflag + "\n\nimport", 1)
		m = "// Auto-modified by gostd2joker at " + curTimeAndVersion() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag)  // [^(]*[(][^)]*[)]
	newImports := "\n"
	importPrefix := "\t_ \"github.com/candid82/joker/std/go/"
	for _, p := range pkgs {
		newImports += importPrefix + p + "\"\n"
	}
	m = reImport.ReplaceAllString(m, flag + newImports + ") // " + endflag)

	if verbose {
		fmt.Printf("Writing %s\n", filepath.ToSlash(f))
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
			fmt.Printf("Adding custom loaded libraries to %s\n", filepath.ToSlash(f))
		}
		m = strings.Replace(m, "\n  *loaded-libs* #{",
			"\n  *loaded-libs* #{\n   ;; " + flag + "\n   ;; " + endflag + "\n", 1)
		m = ";;;; Auto-modified by gostd2joker at " + curTimeAndVersion() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag + "\n *?")
	newImports := "\n  "
	importPrefix := " 'joker.go."
	curLine := ""
	for _, p := range pkgs {
		more := importPrefix + strings.Replace(p, "/", ".", -1)
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
		fmt.Printf("Writing %s\n", filepath.ToSlash(f))
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
			fmt.Printf("Adding custom namespaces to %s\n", filepath.ToSlash(f))
		}
		m = strings.Replace(m, "(def namespaces [",
			"(def namespaces [\n  ;; " + flag + "\n  ;; " + endflag + "\n  ", 1)
		m = ";;;; Auto-modified by gostd2joker at " + curTimeAndVersion() + "\n\n" + m
	}

	reImport := regexp.MustCompile("(?msU)" + flag + ".*" + endflag + "\n *?")
	newImports := "\n "
	importPrefix := " 'go."
	curLine := ""
	for _, p := range pkgs {
		more := importPrefix + strings.Replace(p, "/", ".", -1)
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
		fmt.Printf("Writing %s\n", filepath.ToSlash(f))
	}
	err = ioutil.WriteFile(f, []byte(m), 0777)
	check(err)
}

func packageQuotedImportList(pi packageImports, prefix string, rename bool) string {
	imports := ""
	sortedPackageImports(pi,
		func(k string) {
			if rename {
				imports += prefix + "_" + path.Base(k) + ` "` + k + `"`
			} else {
				imports += prefix + `"` + k + `"`
			}
		})
	return imports
}

func main() {
	fset = token.NewFileSet() // positions are relative to fset
	dump = false

	length := len(os.Args)
	sourceDir := ""
	jokerSourceDir := ""
	replace := false
	overwrite := false

	var mode parser.Mode = parser.ParseComments

	for i := 1; i < length; i++ { // shift
		a := os.Args[i]
		if a[0] == "-"[0] {
			switch a {
			case "--help", "-h":
				usage()
			case "--version", "-V":
				fmt.Printf("%s version %s\n", os.Args[0], VERSION)
			case "--no-timestamp":
				noTimeAndVersion = true
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
		si, e := os.Stat(goLink)
		if e == nil && !si.IsDir() {
			var by []byte
			by, e = ioutil.ReadFile(goLink)
			if e != nil {
				panic("Must specify --go <go-source-dir-name> option, or put <go-source-dir-name> as the first line of a file named ./GO.link")
			}
			m := string(by)
			if idx := strings.IndexAny(m, "\r\n"); idx == -1 {
				goLink = m
			} else {
				goLink = m[0:idx]
			}
			si, e = os.Stat(goLink)
		}
		if e != nil || !si.IsDir() {
			panic(fmt.Sprintf("Must specify --go <go-source-dir-name> option, or make %s a symlink (or text file containing the native path) pointing to the golang/go/ source directory", goLink))
		}
		sourceDir = goLink
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
			if _, e := os.Stat(jokerLibDir); e == nil ||
				(!strings.Contains(e.Error(), "no such file or directory") &&
				!strings.Contains(e.Error(), "The system cannot find the file specified.")) {
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
			func(t string, ti *typeInfo) {
				fmt.Printf("TYPE %s:\n", t)
				fmt.Printf("  %s\n", ti.file)
			})
	}

	/* Generate function code snippets in alphabetical order, to stabilize test output in re unsupported types. */
	sortedFuncInfoMap(unqualifiedFunctions,
		func(f string, v *funcInfo) {
			if v.fd != DUPLICATEFUNCTION {
				genFunction(f, v)
			}
		})

	var out *bufio.Writer
	var unbuf_out *os.File

	sortedPackageMap(jokerCode,
		func(pkgDirUnix string, v codeInfo) {
			pkgBaseName := path.Base(pkgDirUnix)
			if jokerLibDir != "" && jokerLibDir != "-" {
				jf := filepath.Join(jokerLibDir, filepath.FromSlash(pkgDirUnix) + ".joke")
				var e error
				e = os.MkdirAll(filepath.Dir(jf), 0777)
				unbuf_out, e = os.Create(jf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)

				pi := packagesInfo[pkgDirUnix]

				fmt.Fprintf(out,
					`;;;; Auto-generated by gostd2joker at ` + curTimeAndVersion() + `, do not edit!!

(ns
  ^{:go-imports [%s]
    :doc "Provides a low-level interface to the %s package."
    :empty %s}
  go.%s)
`,
					strings.TrimPrefix(packageQuotedImportList(pi.importsAutoGen, " ", false), " "),
					pkgDirUnix,
					func() string {
						if pi.nonEmpty {
							return "false"
						} else {
							return "true"
						}
					}(),
					strings.Replace(pkgDirUnix, "/", ".", -1))
				}
			sortedCodeMap(v,
				func(f string, w string) {
					if verbose || jokerLibDir == "" {
						fmt.Printf("JOKER FUNC %s.%s has:%v\n",
							pkgBaseName, f, w)
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
		func(pkgDirUnix string, v codeInfo) {
			pkgBaseName := path.Base(pkgDirUnix)
			pi := packagesInfo[pkgDirUnix]
			packagesInfo[pkgDirUnix].hasGoFiles = true
			pkgDirNative := filepath.FromSlash(pkgDirUnix)

			if jokerLibDir != "" && jokerLibDir != "-" {
				gf := filepath.Join(jokerLibDir, pkgDirNative,
					pkgBaseName + "_native.go")
				var e error
				e = os.MkdirAll(filepath.Dir(gf), 0777)
				check(e)
				unbuf_out, e = os.Create(gf)
				check(e)
				out = bufio.NewWriterSize(unbuf_out, 16384)

				importCore := ""
				if _, f := pi.importsNative[pkgDirUnix]; f {
					importCore = `
	. "github.com/candid82/joker/core"`
				}

				fmt.Fprintf(out,
					`// Auto-generated by gostd2joker at ` + curTimeAndVersion() + `, do not edit!!

package %s

import (%s%s
)
`,
					pkgBaseName,
					packageQuotedImportList(pi.importsNative, "\n\t", true),
					importCore)
			}
			sortedCodeMap(v,
				func(f string, w string) {
					if verbose || jokerLibDir == "" {
						fmt.Printf("GO FUNC %s.%s has:%v\n",
							pkgBaseName, f, w)
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

		sortedPackagesInfo(packagesInfo,
			func (p string, i *packageInfo) {
				if i.hasGoFiles {
					packagesArray = append(packagesArray, p)
				}
			})
		updateJokerMain(packagesArray, filepath.Join(jokerSourceDir, "main.go"))
		updateCoreDotJoke(packagesArray, filepath.Join(jokerSourceDir, "core", "data", "core.joke"))
		updateGenerateSTD(packagesArray, filepath.Join(jokerSourceDir, "std", "generate-std.joke"))
	}

	if verbose {
		fmt.Printf("Totals: types=%d functions=%d receivers=%d\n",
			len(types), len(unqualifiedFunctions), receivers)
	}

	os.Exit(0)
}

func init() {
	p := `(?m)^(.)`
	nonEmptyLineRegexp = regexp.MustCompile(p)
}
