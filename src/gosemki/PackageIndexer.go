package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	BUILTIN_PKG_CONTENT = `
	type bool bool
	const (
		true  = 0 == 0 // Untyped bool.
		false = 0 != 0 // Untyped bool.
	)
	type uint8 uint8
	type uint16 uint16
	type uint32 uint32
	type uint64 uint64
	type int8 int8
	type int16 int16
	type int32 int32
	type int64 int64
	type float32 float32
	type float64 float64
	type complex64 complex64
	type complex128 complex128
	type string string
	type int int
	type uint uint
	type uintptr uintptr
	type byte byte
	type rune rune
	const iota = 0
	var nil Type
	type Type int
	type Type1 int
	type IntegerType int
	type FloatType float32
	type ComplexType complex64
	func append(slice []Type, elems ...Type) []Type
	func copy(dst, src []Type) int
	func delete(m map[Type]Type1, key Type)
	func len(v Type) int
	func cap(v Type) int
	func make(Type, size IntegerType) Type
	func new(Type) *Type
	func complex(r, i FloatType) ComplexType
	func real(c ComplexType) FloatType
	func imag(c ComplexType) FloatType
	func close(c chan<- Type)
	func panic(v interface{})
	func recover() interface{}
	func print(args ...Type)
	func println(args ...Type)
	type error interface {
		Error() string
	}
	`
)

type PackageIndexer struct {
	fset        *token.FileSet
	files       map[string]*ast.File
	packageName string
	result      *IndexerResult
	lastIdent   *ast.Ident
	context     build.Context
}

func NewPackageIndexer(result *IndexerResult) *PackageIndexer {
	ret := new(PackageIndexer)
	ret.result = result
	return ret
}

func (this *PackageIndexer) ImportPackageScope(path string) (scope *ast.Scope) {
	scope = ast.NewScope(nil)
	// Infering source dir for local imports
	// Not ideal solution since daemon workdir can vary
	srcDir := "."
	if build.IsLocalImport(path) {
		srcDir, _ = os.Getwd()
	}

	// Try to locate package source code and parse it
	pkgInfo, err := build.Import(path, srcDir, build.AllowBinary)
	if err != nil {
		log.Printf("Calling build.Import() for package '%s' failed: %v", path, err)
		return
	}

	nodeMap := make(map[string]*ast.File)
	for _, fileName := range pkgInfo.GoFiles {
		filePath := filepath.Join(pkgInfo.SrcRoot, path, fileName)
		fast, err := parser.ParseFile(this.fset, filePath, nil, parser.ParseComments)
		if fast != nil {
			if ast.FileExports(fast) {
				nodeMap[filePath] = fast
			}
		} else {
			log.Printf("Calling parser.ParseFile() for package '%s' failed: %v", path, err)
			return
		}
	}
	pkgAst, err := ast.NewPackage(this.fset, nodeMap, nil, nil)
	if pkgAst != nil {
		scope = pkgAst.Scope
		exportedNames := make([]string, 0, 1024)
		for key, _ := range scope.Objects {
			exportedNames = append(exportedNames, key)
		}
		log.Printf("Parsed OK package '%s' ast: [%v]", path, exportedNames)
	} else {
		log.Printf("Calling ast.NewPackage() for package '%s' failed: %v", path, err)
		return
	}
	return
}

func (this *PackageIndexer) NewPackage(path string) *ast.Object {
	// TODO: implement unsafe import
	//	if path == "unsafe" {
	//		return types.Unsafe, nil
	//	}

	name := path[strings.LastIndex(path, "/")+1:]
	obj := ast.NewObj(ast.Pkg, name)
	obj.Data = this.ImportPackageScope(path) // required by ast.NewPackage for dot-import
	return obj
}

func (this *PackageIndexer) Import(imports map[string]*ast.Object, path string) (obj *ast.Object, err error) {
	err = nil
	obj = imports[path]
	if obj == nil {
		obj = this.NewPackage(path)
		if obj != nil {
			imports[path] = obj
		}
	}
	return
}

func (this *PackageIndexer) Reindex(filePath string, file []byte) {
	this.packageName = ""
	this.fset = token.NewFileSet()
	this.files = make(map[string]*ast.File)

	this.Parse(filePath, file)
	for _, name := range this.FindAllPackageFiles(filePath) {
		this.Parse(name, nil)
	}
	this.InjectBuiltinPackage()

	pkgAst, errors := ast.NewPackage(this.fset, this.files, this.Import, nil)
	if errors != nil {
		errorList, _ := errors.(scanner.ErrorList)
		this.ParseErrorsInFile(errorList, filePath)
	}
	fileAst := pkgAst.Files[filePath]
	ast.Inspect(fileAst, this.InspectNode)
}

func (this *PackageIndexer) AddIdentRange(ident *ast.Ident) {
	if ident.Obj == nil || ident.Obj.Kind == ast.Bad {
		return
	}

	pos := this.fset.Position(ident.NamePos)
	goRange := GoRange{
		GoPos: GoPos{
			Line:   pos.Line,
			Column: pos.Column,
			Offset: pos.Offset,
		},
		Length: len(ident.Name),
		Kind:   inferIdentKind(ident),
	}
	this.result.AddRange(goRange)
}

func (this *PackageIndexer) AddFuncCallRange(expr *ast.CallExpr) {
	pos := this.fset.Position(expr.Fun.Pos())
	end := this.fset.Position(expr.Fun.End())
	length := pos.Column - end.Column
	if end.Line != pos.Line {
		length = end.Column
	}
	goRange := GoRange{
		GoPos: GoPos{
			Line:   pos.Line,
			Column: pos.Column,
			Offset: pos.Offset,
		},
		Length: length,
		Kind:   GoKindFunc,
	}
	this.result.AddRange(goRange)
}

func (this *PackageIndexer) InspectNode(node ast.Node) bool {
	switch x := node.(type) {
	case *ast.Ident:
		this.AddIdentRange(x)
		return false
	case *ast.CallExpr:
		visitor := CallExprVisitor{indexer: this}
		visitor.ProcessExpr(x)
		return false
	case *ast.KeyValueExpr:
		visitor := KeyValueExprVisitor{indexer: this}
		ast.Inspect(x.Key, visitor.InspectNode)
		visitor.ApplyIdent()
		ast.Inspect(x.Value, this.InspectNode)
		return false
	case *ast.CommentGroup:
		return false
	case *ast.Comment:
		return false
	case *ast.FuncDecl:
		goScope := GoFoldScope{
			LineFrom: this.NodePos(x).Line,
			LineTo:   this.NodeEnd(x).Line,
		}
		this.result.AddFoldScope(goScope)
		pos := this.fset.Position(x.Name.Pos())
		goOutline := GoOutline{
			GoPos: GoPos{
				Line:   pos.Line,
				Column: pos.Column,
				Offset: pos.Offset,
			},
			Name: x.Name.Name,
			Kind: GoKindFunc,
		}
		this.result.AddOutline(goOutline)
		return true
	case *ast.TypeSpec:
		pos := this.fset.Position(x.Name.Pos())
		goOutline := GoOutline{
			GoPos: GoPos{
				Line:   pos.Line,
				Column: pos.Column,
				Offset: pos.Offset,
			},
			Name: x.Name.Name,
			Kind: GoKindType,
		}
		this.result.AddOutline(goOutline)
		return true
	}
	return true
}

func (this *PackageIndexer) NodePos(node ast.Node) token.Position {
	return this.fset.Position(node.Pos())
}

func (this *PackageIndexer) NodeEnd(node ast.Node) token.Position {
	return this.fset.Position(node.End())
}

// Hack to inject `builtin` package definitions into parsed package
func (this *PackageIndexer) InjectBuiltinPackage() {
	var hackContent bytes.Buffer
	hackContent.WriteString("package ")
	hackContent.WriteString(this.packageName)
	hackContent.WriteString(";\n")
	hackContent.WriteString(BUILTIN_PKG_CONTENT)
	this.Parse("", hackContent.Bytes())
}

// Translates *scanner.ErrorList into []GoError
func (this *PackageIndexer) ParseErrorsInFile(errors scanner.ErrorList, filePath string) {
	for _, scanError := range errors {
		if scanError.Pos.Filename == filePath {
			var goerr GoError
			goerr.Line = scanError.Pos.Line
			goerr.Column = scanError.Pos.Column
			goerr.Length = len(scanError.Pos.String())
			goerr.Offset = scanError.Pos.Offset
			goerr.Message = scanError.Msg
			this.result.AddError(goerr)
		}
	}
}

// Finds other files from the same packge as parsed file
func (this *PackageIndexer) FindAllPackageFiles(filePath string) []string {
	dir := path.Dir(filePath)
	file, err := os.Open(dir)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to open sources dir '%s': %s", dir, err.Error())))
	}
	defer file.Close()
	names, err := file.Readdirnames(0)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to read content of sources dir '%s': %s", dir, err.Error())))
	}
	var result []string
	for _, name := range names {
		if strings.HasSuffix(name, ".go") {
			result = append(result, path.Join(dir, name))
		}
	}
	return result
}

func (this *PackageIndexer) Parse(filePath string, src interface{}) {
	fast, err := parser.ParseFile(this.fset, filePath, src, parser.ParseComments)
	if fast == nil {
		panic(errors.New(fmt.Sprintf("Failed to index file, error: '%v'", err)))
	}
	this.files[filePath] = fast
	if len(this.packageName) == 0 {
		this.packageName = fast.Name.Name
	}
}
