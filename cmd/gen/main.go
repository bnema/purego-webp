package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type spec struct {
	Functions []specFunction `json:"functions"`
}

type specFunction struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Symbol    string `json:"symbol,omitempty"`
	Optional  bool   `json:"optional,omitempty"`
}

type tmplFunction struct {
	Name      string
	Signature string
	Symbol    string
	Params    string
	Results   string
	Args      string
	HasReturn bool
	Optional  bool
}

type tmplData struct {
	Functions []tmplFunction
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	specPath := filepath.Join(root, "gen", "spec.json")
	templatePath := filepath.Join(root, "templates", "internal_generated_symbols.go.tmpl")
	outputPath := filepath.Join(root, "internal", "libwebp", "generated_symbols.go")

	sp, err := readSpec(specPath)
	if err != nil {
		panic(err)
	}

	data, err := buildTemplateData(sp)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		panic(err)
	}

	fmted, err := format.Source(out.Bytes())
	if err != nil {
		panic(fmt.Errorf("format generated source: %w\n\n%s", err, out.String()))
	}

	if err := os.WriteFile(outputPath, fmted, 0o644); err != nil {
		panic(err)
	}

	fmt.Println("generated", outputPath)
}

func readSpec(path string) (*spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s spec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func buildTemplateData(s *spec) (*tmplData, error) {
	funcs := make([]tmplFunction, 0, len(s.Functions))
	for _, f := range s.Functions {
		tf, err := parseFunction(f)
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, tf)
	}

	return &tmplData{Functions: funcs}, nil
}

func parseFunction(sf specFunction) (tmplFunction, error) {
	src := fmt.Sprintf("package p\nvar _ %s", sf.Signature)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sig.go", src, 0)
	if err != nil {
		return tmplFunction{}, fmt.Errorf("parse signature for %s: %w", sf.Name, err)
	}

	decl := file.Decls[0].(*ast.GenDecl)
	vs := decl.Specs[0].(*ast.ValueSpec)
	ft := vs.Type.(*ast.FuncType)

	params, args := buildParamsAndArgs(ft.Params)
	results, hasReturn := buildResults(ft.Results)

	return tmplFunction{
		Name:      sf.Name,
		Signature: sf.Signature,
		Symbol:    symbolName(sf),
		Params:    params,
		Results:   results,
		Args:      args,
		HasReturn: hasReturn,
		Optional:  sf.Optional,
	}, nil
}

func symbolName(sf specFunction) string {
	if sf.Symbol != "" {
		return sf.Symbol
	}
	return sf.Name
}

func buildParamsAndArgs(fields *ast.FieldList) (string, string) {
	if fields == nil || len(fields.List) == 0 {
		return "()", "()"
	}

	parts := make([]string, 0, len(fields.List))
	args := make([]string, 0, len(fields.List))
	autoArg := 0

	for _, field := range fields.List {
		typeStr := exprString(field.Type)
		if len(field.Names) == 0 {
			name := fmt.Sprintf("arg%d", autoArg)
			autoArg++
			parts = append(parts, name+" "+typeStr)
			args = append(args, name)
			continue
		}

		names := make([]string, 0, len(field.Names))
		for _, n := range field.Names {
			names = append(names, n.Name)
			args = append(args, n.Name)
		}
		parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
	}

	return "(" + strings.Join(parts, ", ") + ")", "(" + strings.Join(args, ", ") + ")"
}

func buildResults(fields *ast.FieldList) (string, bool) {
	if fields == nil || len(fields.List) == 0 {
		return "", false
	}

	if len(fields.List) == 1 && len(fields.List[0].Names) == 0 {
		return " " + exprString(fields.List[0].Type), true
	}

	parts := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		typeStr := exprString(field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
			continue
		}
		names := make([]string, 0, len(field.Names))
		for _, n := range field.Names {
			names = append(names, n.Name)
		}
		parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
	}

	return " (" + strings.Join(parts, ", ") + ")", true
}

func exprString(expr ast.Expr) string {
	return strings.TrimSpace(typesExprString(expr))
}

func typesExprString(expr ast.Expr) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), expr); err != nil {
		panic(err)
	}
	return b.String()
}
