// https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Stringer is a tool to automate the creation of methods that satisfy the fmt.Stringer
// interface. Given the name of a (signed or unsigned) integer type T that has constants
// defined, stringer will create a new self-contained Go source file implementing
//	func (t T) String() string
// The file is created in the same package and directory as the package that defines T.
// It has helpful defaults designed for use with go generate.
//
// Stringer works best with constants that are consecutive values such as created using iota,
// but creates good code regardless. In the future it might also provide custom support for
// constant sets that are bit patterns.
//
// For example, given this snippet,
//
//	package painkiller
//
//	type Pill int
//
//	const (
//		Placebo Pill = iota
//		Aspirin
//		Ibuprofen
//		Paracetamol
//		Acetaminophen = Paracetamol
//	)
//
// running this command
//
//	stringer -type=Pill
//
// in the same directory will create the file pill_string.go, in package painkiller,
// containing a definition of
//
//	func (Pill) String() string
//
// That method will translate the value of a Pill constant to the string representation
// of the respective constant name, so that the call fmt.Print(painkiller.Aspirin) will
// print the string "Aspirin".
//
// Typically this process would be run using go generate, like this:
//
//	//go:generate stringer -type=Pill
//
// If multiple constants have the same value, the lexically first matching name will
// be used (in the example, Acetaminophen will print as "Paracetamol").
//
// With no arguments, it processes the package in the current directory.
// Otherwise, the arguments must name a single directory holding a Go package
// or a set of Go source files that represent a single Go package.
//
// The -type flag accepts a comma-separated list of types so a single run can
// generate methods for multiple types. The default output file is t_string.go,
// where t is the lower-cased name of the first type listed. It can be overridden
// with the -output flag.
//
// The -linecomment flag tells stringer to generate the text of any line comment, trimmed
// of leading spaces, instead of the constant name. For instance, if the constants above had a
// Pill prefix, one could write
//
//	PillAspirin // Aspirin
//
// to suppress it in the output.
package main // import "golang.org/x/tools/cmd/stringer"

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	DefCodeIndex  = "CodeIndex"
	DefNameIndex  = "NameIndex"
	DefCodeMap    = "CodeMap"
	DefNameMap    = "NameMap"
	DefCode2IDMap = "Code2IDMap"
	DefCodeVal    = "CodeName"
	DefNameVal    = "Name"
	DefCodeFn     = "Code"
	DefNameFn     = "Name"
	DefCode2IDFn  = "CodeTo"
)

var (
	typeNames     = flag.String("type", "", "comma-separated list of type names; must be set")
	output        = flag.String("output", "", "output file name; default srcdir/<type>_string.go")
	buildTags     = flag.String("tags", "", "comma-separated list of build tags to apply")
	codeFnName    = flag.String("code", "Code", "code函数名")
	nameFnName    = flag.String("name", "Name", "name函数名")
	code2IDFnName = flag.String("code2id", "", "code转id函数名")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of stringer:\n")
	fmt.Fprintf(os.Stderr, "\tstringer [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tstringer [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://pkg.go.dev/golang.org/x/tools/cmd/stringer\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("stringer: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	types := strings.Split(*typeNames, ",")
	var tags []string
	if len(*buildTags) > 0 {
		tags = strings.Split(*buildTags, ",")
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	// Parse the package once.
	var dir string
	g := Generator{
		codeFnName:    *codeFnName,
		nameFnName:    *nameFnName,
		code2IDFnName: *code2IDFnName,
	}
	g.codeFnName = *codeFnName
	if g.codeFnName == "" {
		g.codeFnName = DefCodeFn
	}
	g.nameFnName = *nameFnName
	if g.nameFnName == "" {
		g.nameFnName = DefNameFn
	}

	// TODO(suzmue): accept other patterns for packages (directories, list of files, import paths, etc).
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		if len(tags) != 0 {
			log.Fatal("-tags option applies only to directories, not when files are specified")
		}
		dir = filepath.Dir(args[0])
	}

	g.parsePackage(args, tags)

	// Print the header and package clause.
	g.Printf("// Code generated by \"stringer %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	g.Printf("\n")
	g.Printf("package %s", g.pkg.name)
	g.Printf("\n")
	g.Printf("import \"strconv\"\n") // Used by all methods.

	// Run generate for each type.
	for _, typeName := range types {
		g.generate(typeName)
		g.Printf("\n")
	}

	// Format the output.
	src := g.format()

	// Write to file.
	outputName := *output
	if outputName == "" {
		baseName := fmt.Sprintf("%s_string.go", types[0])
		outputName = filepath.Join(dir, strings.ToLower(baseName))
	}
	err := ioutil.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf bytes.Buffer // Accumulated output.
	pkg *Package     // Package we are scanning.

	codeFnName    string
	nameFnName    string
	code2IDFnName string
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// File holds a single parsed file and associated data.
type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
	// These fields are reset for each type being generated.
	typeName string  // Name of the constant type.
	values   []Value // Accumulator for constant values of that type.
}

type Package struct {
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

// parsePackage analyzes the single package constructed from the patterns and tags.
// parsePackage exits if there is an error.
func (g *Generator) parsePackage(patterns []string, tags []string) {
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

// generate produces the String method for the named type.
func (g *Generator) generate(typeName string) {
	values := make([]Value, 0, 100)
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		file.typeName = typeName
		file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, file.genDecl)
			values = append(values, file.values...)
		}
	}

	if len(values) == 0 {
		log.Fatalf("no values defined for type %s", typeName)
	}
	// Generate code that will fail if the constants change value.
	g.Printf("func _() {\n")
	g.Printf("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
	g.Printf("\t// Re-run the stringer command to generate them again.\n")
	g.Printf("\tvar x [1]struct{}\n")
	for _, v := range values {
		g.Printf("\t_ = x[%s - %s]\n", v.originalName, v.str)
	}
	g.Printf("}\n")
	runs := splitIntoRuns(values)
	// The decision of which pattern to use depends on the number of
	// runs in the numbers. If there's only one, it's easy. For more than
	// one, there's a tradeoff between complexity and size of the data
	// and code vs. the simplicity of a map. A map takes more space,
	// but so does the code. The decision here (crossover at 10) is
	// arbitrary, but considers that for large numbers of runs the cost
	// of the linear scan in the switch might become important, and
	// rather than use yet another algorithm such as binary search,
	// we punt and use a map. In any case, the likelihood of a map
	// being necessary for any realistic example other than bitmasks
	// is very low. And bitmasks probably deserve their own analysis,
	// to be done some other day.
	switch {
	case len(runs) == 1:
		g.buildOneRun(runs, typeName)
		g.code2ID(runs, typeName)
	case len(runs) <= 10:
		g.buildMultipleRuns(runs, typeName)
		g.code2ID2(runs, typeName)
	default:
		g.buildMap(runs, typeName)
		g.code2ID(runs, typeName)
	}
}

// splitIntoRuns breaks the values into runs of contiguous sequences.
// For example, given 1,2,3,5,6,7 it returns {1,2,3},{5,6,7}.
// The input slice is known to be non-empty.
func splitIntoRuns(values []Value) [][]Value {
	// We use stable sort so the lexically first name is chosen for equal elements.
	sort.Stable(byValue(values))
	// Remove duplicates. Stable sort has put the one we want to print first,
	// so use that one. The String method won't care about which named constant
	// was the argument, so the first name for the given value is the only one to keep.
	// We need to do this because identical values would cause the switch or map
	// to fail to compile.
	j := 1
	for i := 1; i < len(values); i++ {
		if values[i].value != values[i-1].value {
			values[j] = values[i]
			j++
		}
	}
	values = values[:j]
	runs := make([][]Value, 0, 10)
	for len(values) > 0 {
		// One contiguous sequence per outer loop.
		i := 1
		for i < len(values) && values[i].value == values[i-1].value+1 {
			i++
		}
		runs = append(runs, values[:i])
		values = values[i:]
	}
	return runs
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

// Value represents a declared constant.
type Value struct {
	originalName string // The name of the constant.
	codeName     string // The name with trimmed prefix.
	cnName       string
	// The value is stored as a bit pattern alone. The boolean tells us
	// whether to interpret it as an int64 or a uint64; the only place
	// this matters is when sorting.
	// Much of the time the str field is all we need; it is printed
	// by Value.String.
	value  uint64 // Will be converted to int64 when needed.
	signed bool   // Whether the constant is a signed type.
	str    string // The string representation given by the "go/constant" package.
}

func (v *Value) String() string {
	return v.str
}

func ValueCode(v *Value) string {
	return v.codeName
}

func ValueName(v *Value) string {
	return v.cnName
}

// byValue lets us sort the constants into increasing order.
// We take care in the Less method to sort in signed or unsigned order,
// as appropriate.
type byValue []Value

func (b byValue) Len() int      { return len(b) }
func (b byValue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byValue) Less(i, j int) bool {
	if b[i].signed {
		return int64(b[i].value) < int64(b[j].value)
	}
	return b[i].value < b[j].value
}

// genDecl processes one declaration clause.
func (f *File) genDecl(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST {
		// We only care about const declarations.
		return true
	}
	// The name of the type of the constants we are declaring.
	// Can change if this is a multi-element declaration.
	typ := ""
	// Loop over the elements of the declaration. Each element is a ValueSpec:
	// a list of names possibly followed by a type, possibly followed by values.
	// If the type and value are both missing, we carry down the type (and value,
	// but the "go/types" package takes care of that).
	for _, spec := range decl.Specs {
		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
		if vspec.Type == nil && len(vspec.Values) > 0 {
			// "X = 1". With no type but a value. If the constant is untyped,
			// skip this vspec and reset the remembered type.
			typ = ""

			// If this is a simple type conversion, remember the type.
			// We don't mind if this is actually a call; a qualified call won't
			// be matched (that will be SelectorExpr, not Ident), and only unusual
			// situations will result in a function call that appears to be
			// a type conversion.
			ce, ok := vspec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			id, ok := ce.Fun.(*ast.Ident)
			if !ok {
				continue
			}
			typ = id.Name
		}
		if vspec.Type != nil {
			// "X T". We have a type. Remember it.
			ident, ok := vspec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			typ = ident.Name
		}
		if typ != f.typeName {
			// This is not the type we're looking for.
			continue
		}
		// We now have a list of names (from one line of source code) all being
		// declared with the desired type.
		// Grab their names and actual values and store them in f.values.
		for _, name := range vspec.Names {
			if name.Name == "_" {
				continue
			}
			// This dance lets the type checker find the values for us. It's a
			// bit tricky: look up the object declared by the name, find its
			// types.Const, and extract its value.
			obj, ok := f.pkg.defs[name]
			if !ok {
				log.Fatalf("no value for constant %s", name)
			}
			info := obj.Type().Underlying().(*types.Basic).Info()
			if info&types.IsInteger == 0 {
				log.Fatalf("can't handle non-integer constant type %s", typ)
			}
			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
			if value.Kind() != constant.Int {
				log.Fatalf("can't happen: constant is not an integer %s", name)
			}
			i64, isInt := constant.Int64Val(value)
			u64, isUint := constant.Uint64Val(value)
			if !isInt && !isUint {
				log.Fatalf("internal error: value of %s is not an integer: %s", name, value.String())
			}
			if !isInt {
				u64 = uint64(i64)
			}
			v := Value{
				originalName: name.Name,
				value:        u64,
				signed:       info&types.IsUnsigned == 0,
				str:          value.String(),
			}
			if c := vspec.Comment; c != nil && len(c.List) == 1 {
				r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
				names := r.FindAllString(strings.TrimSpace(c.Text()), -1)
				if len(names) > 0 {
					v.codeName = strings.Trim(names[0], "\"")
				}
				if len(names) > 1 {
					v.cnName = strings.Trim(names[1], "\"")
				}
			}
			if v.cnName == "" {
				v.cnName = v.originalName
			}
			if v.codeName == "" {
				v.codeName = v.originalName
			}
			f.values = append(f.values, v)
		}
	}
	return false
}

// Helpers

// usize returns the number of bits of the smallest unsigned integer
// type that will hold n. Used to create the smallest possible slice of
// integers to use as indexes into the concatenated strings.
func usize(n int) int {
	switch {
	case n < 1<<8:
		return 8
	case n < 1<<16:
		return 16
	default:
		// 2^32 is enough constants for anyone.
		return 32
	}
}

// declareIndexAndNameVars declares the index slices and concatenated names
// strings representing the runs of values.
func (g *Generator) declareIndexAndNameVars(runs [][]Value, typeName string) {
	var indexes, names []string
	for i, run := range runs {
		indexs, namex := g.createIndexAndNameDecl(run, typeName, fmt.Sprintf("_%d", i))
		if len(run) != 1 {
			indexes = append(indexes, indexs[0], indexs[1])
		}
		names = append(names, namex[0], namex[1])
	}
	g.Printf("const (\n")
	for _, name := range names {
		g.Printf("\t%s\n", name)
	}
	g.Printf(")\n\n")

	if len(indexes) > 0 {
		g.Printf("var (")
		for _, index := range indexes {
			g.Printf("\t%s\n", index)
		}
		g.Printf(")\n\n")
	}
}

// declareIndexAndNameVar is the single-run version of declareIndexAndNameVars
func (g *Generator) declareIndexAndNameVar(run []Value, typeName string) {
	indexs, names := g.createIndexAndNameDecl(run, typeName, "")
	g.Printf("const (\n")
	g.Printf("%s\n", names[0])
	g.Printf("%s\n", names[1])
	g.Printf(")\n\n")

	g.Printf("var (\n")
	g.Printf("%s\n", indexs[0])
	g.Printf("%s\n", indexs[1])
	g.Printf(")\n\n")
}

// createIndexAndNameDecl returns the pair of declarations for the run. The caller will add "const" and "var".
func (g *Generator) createIndexAndNameDecl(
	run []Value,
	typeName string,
	suffix string,
) ([2]string, [2]string) {
	f := func(nameKey, indexKey string, fn func(*Value) string) (string, string) {
		b := new(bytes.Buffer)
		indexes := make([]int, len(run))
		for i := range run {
			b.WriteString(fn(&run[i]))
			indexes[i] = b.Len()
		}
		nameConst := fmt.Sprintf("_%s%s%s = %q", typeName, nameKey, suffix, b.String())
		nameLen := b.Len()

		b.Reset()
		fmt.Fprintf(b, "_%s%s%s = [...]uint%d{0, ", typeName, indexKey, suffix, usize(nameLen))
		for i, v := range indexes {
			if i > 0 {
				fmt.Fprintf(b, ", ")
			}
			fmt.Fprintf(b, "%d", v)
		}
		fmt.Fprintf(b, "}")
		return b.String(), nameConst
	}

	idx1, name1 := f(DefCodeVal, DefCodeIndex, ValueCode)
	idx2, name2 := f(DefNameVal, DefNameIndex, ValueName)
	return [2]string{idx1, idx2}, [2]string{name1, name2}
}

// declareNameVars declares the concatenated names string representing all the values in the runs.
func (g *Generator) declareNameVars(runs [][]Value, typeName string, suffix string) {
	g.Printf("const (\n")
	f := func(nameKey string, fn func(*Value) string) {
		g.Printf("\t_%s%s%s = \"", typeName, nameKey, suffix)
		for _, run := range runs {
			for i := range run {
				g.Printf("%s", fn(&run[i]))
			}
		}
		g.Printf("\"\n")
	}
	f(DefCodeVal, ValueCode)
	f(DefNameVal, ValueName)
	g.Printf(")\n")
}

// buildOneRun generates the variables and String method for a single run of contiguous values.
func (g *Generator) buildOneRun(runs [][]Value, typeName string) {
	values := runs[0]
	g.Printf("\n")
	g.declareIndexAndNameVar(values, typeName)
	// The generated code is simple enough to write as a Printf format.
	lessThanZero := ""
	if values[0].signed {
		lessThanZero = "i < 0 || "
	}
	if values[0].value == 0 { // Signed or unsigned, 0 is still 0.
		g.Printf(
			stringOneRun, typeName, usize(len(values)), lessThanZero,
			g.codeFnName, DefCodeVal, DefCodeIndex,
		)
		g.Printf("\n")
		g.Printf(
			stringOneRun, typeName, usize(len(values)), lessThanZero,
			g.nameFnName, DefNameVal, DefNameIndex,
		)
	} else {
		g.Printf(
			stringOneRunWithOffset, typeName, values[0].String(), usize(len(values)),
			lessThanZero, g.codeFnName, DefCodeVal, DefCodeIndex,
		)
		g.Printf("\n")
		g.Printf(
			stringOneRunWithOffset, typeName, values[0].String(), usize(len(values)),
			lessThanZero, g.nameFnName, DefNameVal, DefNameIndex,
		)
	}
}

// Arguments to format are:
//	[1]: type name
//	[2]: size of index element (8 for uint8 etc.)
//	[3]: less than zero check (for signed types)
const stringOneRun = `func (i %[1]s) %[4]s() string {
	if %[3]si >= %[1]s(len(_%[1]s%[6]s)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _%[1]s%[5]s[_%[1]s%[6]s[i]:_%[1]s%[6]s[i+1]]
}
`

// Arguments to format are:
//	[1]: type name
//	[2]: lowest defined value for type, as a string
//	[3]: size of index element (8 for uint8 etc.)
//	[4]: less than zero check (for signed types)
/*
 */
const stringOneRunWithOffset = `func (i %[1]s) %[5]s() string {
	i -= %[2]s
	if %[4]si >= %[1]s(len(_%[1]s%[7]s)-1) {
		return "%[1]s(" + strconv.FormatInt(int64(i + %[2]s), 10) + ")"
	}
	return _%[1]s%[6]s[_%[1]s%[7]s[i] : _%[1]s%[7]s[i+1]]
}
`

// buildMultipleRuns generates the variables and String method for multiple runs of contiguous values.
// For this pattern, a single Printf format won't do.
func (g *Generator) buildMultipleRuns(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.declareIndexAndNameVars(runs, typeName)

	f := func(funcName, nameKey, indexKey string) {
		g.Printf("func (i %s) %s() string {\n", typeName, funcName)
		g.Printf("\tswitch {\n")
		for i, values := range runs {
			if len(values) == 1 {
				g.Printf("\tcase i == %s:\n", &values[0])
				g.Printf("\t\treturn _%s%s_%d\n", typeName, nameKey, i)
				continue
			}
			if values[0].value == 0 && !values[0].signed {
				// For an unsigned lower bound of 0, "0 <= i" would be redundant.
				g.Printf("\tcase i <= %s:\n", &values[len(values)-1])
			} else {
				g.Printf("\tcase %s <= i && i <= %s:\n", &values[0], &values[len(values)-1])
			}
			if values[0].value != 0 {
				g.Printf("\t\ti -= %s\n", &values[0])
			}
			g.Printf(
				"\t\treturn _%s%s_%d[_%s%s_%d[i]:_%s%s_%d[i+1]]\n",
				typeName, nameKey, i, typeName, indexKey, i, typeName, indexKey, i,
			)
		}
		g.Printf("\tdefault:\n")
		g.Printf("\t\treturn \"%s(\" + strconv.FormatInt(int64(i), 10) + \")\"\n", typeName)
		g.Printf("\t}\n")
		g.Printf("}\n")
	}
	f(g.codeFnName, DefCodeVal, DefCodeIndex)
	g.Printf("\n")
	f(g.nameFnName, DefNameVal, DefNameIndex)
}

// buildMap handles the case where the space is so sparse a map is a reasonable fallback.
// It's a rare situation but has simple code.
func (g *Generator) buildMap(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.declareNameVars(runs, typeName, "")
	f := func(mapName, nameKey string, fn func(*Value) string) {
		g.Printf("\nvar _%s%s = map[%s]string{\n", typeName, mapName, typeName)
		n := 0
		for _, values := range runs {
			for _, value := range values {
				g.Printf("\t%s: _%s%s[%d:%d],\n", &value, typeName, nameKey, n, n+len(fn(&value)))
				n += len(fn(&value))
			}
		}
		g.Printf("}\n\n")
	}
	f(DefCodeMap, DefCodeVal, ValueCode)
	f(DefNameMap, DefNameVal, ValueName)
	g.Printf(stringMap, typeName, g.codeFnName, DefCodeMap)
	g.Printf("\n")
	g.Printf(stringMap, typeName, g.nameFnName, DefNameMap)
}

// Argument to format is the type name.
const stringMap = `func (i %[1]s) %[2]s() string {
	if str, ok := _%[1]s%[3]s[i]; ok {
		return str
	}
	return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
}
`

func (g *Generator) code2ID(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.Printf("\nvar _%s%s = map[string]%s{\n", typeName, DefCode2IDMap, typeName)
	n := 0
	for _, values := range runs {
		for _, value := range values {
			g.Printf("\t_%s%s[%d:%d]: %s,\n", typeName, DefCodeVal, n, n+len(ValueCode(&value)), &value)
			n += len(ValueCode(&value))
		}
	}

	fnName := g.code2IDFnName
	if fnName == "" {
		fnName = fmt.Sprintf("%s%s", DefCode2IDFn, typeName)
	}

	g.Printf("}\n\n")
	g.Printf(stringCode2IDMap, typeName, fnName, DefCode2IDMap)
	g.Printf("\n")
}

func (g *Generator) code2ID2(runs [][]Value, typeName string) {
	g.Printf("\n")
	g.Printf("\nvar _%s%s = map[string]%s{\n", typeName, DefCode2IDMap, typeName)
	n := 0
	for i, values := range runs {
		for _, value := range values {
			g.Printf("\t_%s%s_%d: %s,\n", typeName, DefCodeVal, i, &value)
			n += len(ValueCode(&value))
		}
	}

	fnName := g.code2IDFnName
	if fnName == "" {
		fnName = fmt.Sprintf("%s%s", DefCode2IDFn, typeName)
	}

	g.Printf("}\n\n")
	g.Printf(stringCode2IDMap, typeName, fnName, DefCode2IDMap)
	g.Printf("\n")
}

const stringCode2IDMap = `func %[2]s(code string, dftVal %[1]s) %[1]s {
	if val, ok := _%[1]s%[3]s[code]; ok {
		return val
	}
	return dftVal
}
`
