package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"strconv"
)

const (
	GoKindBad = iota
	GoKindPkg
	GoKindConst
	GoKindType
	GoKindVar
	GoKindField
	GoKindFunc
	GoKindLabel
)

func isAstObjectAField(obj *ast.Object) bool {
	if obj.Decl == nil {
		return false
	}
	switch obj.Decl.(type) {
	case *ast.Field:
		return true
	}
	return false
}

func inferIdentKind(ident *ast.Ident) int {
	switch ident.Obj.Kind {
	case ast.Pkg:
		return GoKindPkg
	case ast.Con:
		return GoKindConst
	case ast.Typ:
		return GoKindType
	case ast.Var:
		if isAstObjectAField(ident.Obj) {
			return GoKindField
		}
		return GoKindVar
	case ast.Fun:
		return GoKindFunc
	case ast.Lbl:
		return GoKindLabel
	}
	return GoKindBad
}

func goKindToString(kind int) string {
	switch kind {
	case int(GoKindPkg):
		return "pkg"
	case int(GoKindConst):
		return "con"
	case int(GoKindType):
		return "typ"
	case int(GoKindVar):
		return "var"
	case int(GoKindField):
		return "fld"
	case int(GoKindFunc):
		return "fun"
	case int(GoKindLabel):
		return "lbl"
	}
	return ""
}

type GoPos struct {
	Line   int
	Column int
	Offset int
}

type GoRange struct {
	GoPos
	Length int
	Kind   int // ast.ObjKind transformed to int
}

type GoOutline struct {
	GoPos
	Name string
	Kind int // ast.ObjKind transformed to int
}

type GoError struct {
	GoPos
	Length  int
	Message string
}

type GoFoldScope struct {
	LineFrom int
	LineTo   int
}

type IndexerResult struct {
	Ranges  []GoRange
	Errors  []GoError
	Folds   []GoFoldScope
	Outline []GoOutline
	InPanic bool
}

func (this *GoRange) MarshalJSON() ([]byte, error) {
	var jsonBytes bytes.Buffer
	jsonBytes.WriteString("{\"lin\":")
	jsonBytes.WriteString(strconv.Itoa(this.Line))
	jsonBytes.WriteString(",\"col\":")
	jsonBytes.WriteString(strconv.Itoa(this.Column))
	jsonBytes.WriteString(",\"off\":")
	jsonBytes.WriteString(strconv.Itoa(this.Offset))
	jsonBytes.WriteString(",\"len\":")
	jsonBytes.WriteString(strconv.Itoa(this.Length))
	jsonBytes.WriteString(",\"knd\":\"")
	jsonBytes.WriteString(goKindToString(this.Kind))
	jsonBytes.WriteString("\"}")
	return jsonBytes.Bytes(), nil
}

func (this *GoOutline) MarshalJSON() ([]byte, error) {
	// TODO: escape this.Name quotes
	var jsonBytes bytes.Buffer
	jsonBytes.WriteString("{\"lin\":")
	jsonBytes.WriteString(strconv.Itoa(this.Line))
	jsonBytes.WriteString(",\"col\":")
	jsonBytes.WriteString(strconv.Itoa(this.Column))
	jsonBytes.WriteString(",\"off\":")
	jsonBytes.WriteString(strconv.Itoa(this.Offset))
	jsonBytes.WriteString(",\"str\":\"")
	jsonBytes.WriteString(this.Name)
	jsonBytes.WriteString("\",\"knd\":\"")
	jsonBytes.WriteString(goKindToString(this.Kind))
	jsonBytes.WriteString("\"}")
	return jsonBytes.Bytes(), nil
}

func (this *GoError) MarshalJSON() ([]byte, error) {
	var jsonBytes bytes.Buffer
	jsonBytes.WriteString("{\"lin\":")
	jsonBytes.WriteString(strconv.Itoa(this.Line))
	jsonBytes.WriteString(",\"col\":")
	jsonBytes.WriteString(strconv.Itoa(this.Column))
	jsonBytes.WriteString(",\"off\":")
	jsonBytes.WriteString(strconv.Itoa(this.Offset))
	jsonBytes.WriteString(",\"len\":")
	jsonBytes.WriteString(strconv.Itoa(this.Length))
	jsonBytes.WriteString(",\"msg\":\"")
	// TODO: escape this.Message quotes
	jsonBytes.WriteString(this.Message)
	jsonBytes.WriteString("\"}")
	return jsonBytes.Bytes(), nil
}

func (this *GoFoldScope) MarshalJSON() ([]byte, error) {
	var jsonBytes bytes.Buffer
	jsonBytes.WriteString("{\"from\":")
	jsonBytes.WriteString(strconv.Itoa(this.LineFrom))
	jsonBytes.WriteString(",\"to\":")
	jsonBytes.WriteString(strconv.Itoa(this.LineTo))
	jsonBytes.WriteString("}")
	return jsonBytes.Bytes(), nil
}

func (this *IndexerResult) MarshalJSON() ([]byte, error) {
	rangesStr, _ := json.Marshal(this.Ranges)
	errorsStr, _ := json.Marshal(this.Errors)
	foldsStr, _ := json.Marshal(this.Folds)
	outlineStr, _ := json.Marshal(this.Outline)

	var jsonBytes bytes.Buffer
	jsonBytes.WriteString("{\"ranges\":")
	jsonBytes.Write(rangesStr)
	jsonBytes.WriteString("{\"outline\":")
	jsonBytes.Write(outlineStr)
	jsonBytes.WriteString(",\"errors\":")
	jsonBytes.Write(errorsStr)
	jsonBytes.WriteString(",\"folds\":")
	jsonBytes.Write(foldsStr)
	jsonBytes.WriteString(",\"in_panic\":")
	if this.InPanic {
		jsonBytes.WriteString("true")
	} else {
		jsonBytes.WriteString("false")
	}
	jsonBytes.WriteString("\"}")
	return jsonBytes.Bytes(), nil
}

func (this *IndexerResult) AddRange(goRange GoRange) {
	this.Ranges = append(this.Ranges, goRange)
}

func (this *IndexerResult) AddOutline(goOutline GoOutline) {
	this.Outline = append(this.Outline, goOutline)
}

func (this *IndexerResult) AddError(goError GoError) {
	this.Errors = append(this.Errors, goError)
}

func (this *IndexerResult) AddFoldScope(goScope GoFoldScope) {
	this.Folds = append(this.Folds, goScope)
}
