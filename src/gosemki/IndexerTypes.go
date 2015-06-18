package main

import (
    "go/ast"
    "bytes"
    "strconv"
    "encoding/json"
)

type GoPos struct {
    Line int
    Column int
    Offset int
}

type GoRange struct {
    GoPos
    Length int
    Kind ast.ObjKind
}

type GoError struct {
    GoPos
    Length int
    Message string
}

type GoFoldScope struct {
    LineFrom int
    LineTo int
}

type IndexerResult struct {
    Ranges  []GoRange
    Errors  []GoError
    Folds   []GoFoldScope
}

func (this *GoRange) MarshalJSON() ([]byte, error) {
    var kind string
    switch this.Kind {
    case ast.Pkg:
        kind = "pkg"
    case ast.Con:
        kind = "con"
    case ast.Typ:
        kind = "typ"
    case ast.Var:
        kind = "var"
    case ast.Fun:
        kind = "fun"
    case ast.Lbl:
        kind = "lbl"
    }
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
    jsonBytes.WriteString(kind)
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
    jsonBytes.WriteString("\"}")
    return jsonBytes.Bytes(), nil
}

func (this *IndexerResult) MarshalJSON() ([]byte, error) {
    rangesStr, _ := json.Marshal(this.Ranges)
    errorsStr, _ := json.Marshal(this.Errors)

    var jsonBytes bytes.Buffer
    jsonBytes.WriteString("{\"ranges\":")
    jsonBytes.Write(rangesStr)
    jsonBytes.WriteString(",\"errors\":")
    jsonBytes.Write(errorsStr)
    jsonBytes.WriteString("\"}")
    return jsonBytes.Bytes(), nil
}
