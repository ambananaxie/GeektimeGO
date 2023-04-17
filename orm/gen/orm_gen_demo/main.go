package main

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"text/template"
)

//go:embed tpl.gohtml
var genOrm string
// 调用这个方法来生成代码
func gen(w io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, Data{
		File: file,
		Ops: []string{"LT", "GT", "EQ"},
	})
}

type Data struct {
	*File
	Ops []string
}