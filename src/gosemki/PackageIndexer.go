package main

import (
    "go/parser"
    "go/token"
    "go/ast"
    "path"
    "os"
    "strings"
    "go/scanner"
    "bytes"
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

type GoRange struct {
    Line int
    Column int
    Offset int
    Length int
    Kind ast.ObjKind
}
type GoError struct {
    Line int
    Column int
    Offset int
    Length int
    Message string
}

type PackageIndexer struct {
    fset            *token.FileSet
    files           map[string]*ast.File
    errors          []GoError
    ranges          []GoRange
    packageName     string
}

func DefaultImporter(imports map[string]*ast.Object, path string) (*ast.Object, error) {
    pkg := imports[path]
    if pkg == nil {
        name := path[strings.LastIndex(path, "/")+1:]
        pkg = ast.NewObj(ast.Pkg, name)
        pkg.Data = ast.NewScope(nil) // required by ast.NewPackage for dot-import
        imports[path] = pkg
    }
    imports[path] = pkg
    return pkg, nil
}

func (this *PackageIndexer) Reindex(filePath string, file []byte) {
    this.packageName = ""
    this.fset = token.NewFileSet()
    this.files = make(map[string]*ast.File)
    this.errors = []GoError{}
    this.ranges = []GoRange{}

    this.Parse(filePath, file)
    for _, name := range this.FindAllPackageFiles(filePath) {
        this.Parse(name, nil)
    }
    this.InjectBuiltinPackage()

    pkgAst, errors := ast.NewPackage(this.fset, this.files, DefaultImporter, nil)
    if errors != nil {
        errorList, _ := errors.(scanner.ErrorList)
        this.ParseErrorsInFile(errorList, filePath)
    }
    fileAst := pkgAst.Files[filePath]
    ast.Inspect(fileAst, func (node ast.Node) bool {
        return this.InspectNode(node)
    })
}

func (this *PackageIndexer) InspectNode(node ast.Node) bool {
    switch x := node.(type) {
        case *ast.Ident:
            if x.Obj != nil && x.Obj.Kind != ast.Bad {
                pos := this.fset.Position(x.NamePos)
                goRange := GoRange{
                    Line:   pos.Line,
                    Column: pos.Column,
                    Offset: pos.Offset,
                    Length: len(x.Name),
                    Kind:   x.Obj.Kind,
                }
                this.ranges = append(this.ranges, goRange)
            }
            return false
        case *ast.CommentGroup:
            return false
        case *ast.Comment:
            return false
        case *ast.CompositeLit:
            return false
    }
    return true
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

func (this *PackageIndexer) ParseErrorsInFile(errors scanner.ErrorList, filePath string) {
    for _, scanError := range errors {
        if scanError.Pos.Filename == filePath {
            var goerr GoError
            goerr.Line = scanError.Pos.Line
            goerr.Column = scanError.Pos.Column
            goerr.Length = len(scanError.Pos.String())
            goerr.Offset = scanError.Pos.Offset
            goerr.Message = scanError.Msg
            this.errors = append(this.errors, goerr)
        }
    }
}

func (this *PackageIndexer) FindAllPackageFiles(filePath string) []string {
    dir := path.Dir(filePath)
    file, err := os.Open(dir)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    names, err := file.Readdirnames(0)
    if err != nil {
        panic(err)
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
    if err != nil {
        panic(err)
    }
    this.files[filePath] = fast
    if len(this.packageName) == 0 {
        this.packageName = fast.Name.Name
    }
}
