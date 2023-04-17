package main

import (
	"go/ast"
)

type SingleFileEntryVisitor struct {
	file *FileVisitor
}

func (s *SingleFileEntryVisitor) Get() *File {
	types := make([]Type, 0, len(s.file.types))
	for _, typ := range s.file.types {
		types = append(types, Type{
			Name: typ.name,
			Fields: typ.fields,
		})
	}
	return &File{
		Package: s.file.Package,
		Imports: s.file.Imports,
		Types: types,
	}
}

func (s *SingleFileEntryVisitor) Visit(node ast.Node) (w ast.Visitor) {
	fn, ok := node.(*ast.File)
	if !ok {
		// 不是我们要的代表文件的节点
		return s
	}
	// fn.Name 就是包名
	s.file = &FileVisitor{
		Package: fn.Name.String(),
	}
	return s.file
}

type File struct {
	Package string
	Imports []string
	Types []Type
}

type FileVisitor struct {
	Package string
	Imports []string
	types []*TypeVisitor
}

func (f *FileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.TypeSpec:
		v := &TypeVisitor{name: n.Name.String()}
		f.types = append(f.types, v)
		return v
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			path = n.Name.String() + " " + path
		}
		f.Imports = append(f.Imports, path)
	}
	return f
}

type TypeVisitor struct {
	name string
	fields []Field
}

func (t *TypeVisitor) Visit(node ast.Node) (w ast.Visitor) {
	n, ok := node.(*ast.Field)
	if !ok {
		return t
	}
	var typ string
	switch nt := n.Type.(type) {
	case *ast.Ident:
		typ = nt.String()
	case *ast.StarExpr:
		switch xt := nt.X.(type) {
		case *ast.Ident:
			typ = "*" + xt.String()
		case *ast.SelectorExpr:
			typ = "*" + xt.X.(*ast.Ident).String() + "." + xt.Sel.String()
		}
	case *ast.ArrayType:
		typ = "[]byte"
	default:
		panic("不支持的类型")
	}
	for _, name := range n.Names {
		t.fields = append(t.fields, Field{
			Name: name.String(),
			Type: typ,
		})
	}
	return t
}

type Type struct {
	Name string
	Fields []Field
}


type Field struct {
	Name string
	Type string
}

