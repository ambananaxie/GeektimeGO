package web

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	// Render 渲染页面
	// tplName 模板的名字，按名索引
	// data 渲染页面用的数据
	Render(ctx context.Context, tplName string, data any) ([]byte, error)

	// 渲染页面，数据写入到 writer 里面
	// Render(ctx, "aa", map[]{}, repsonseWriter)
	// Render(ctx context.Context, tplName string, data any, writer io.Writer) error

	// 不需要，让具体实现自己管自己的模板
	// AddTemplate(tlpName string, tpl []byte) error

	// 用这个Context 没有问题
	// Render(ctx Context)
}

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}

func (g *GoTemplateEngine) ParseGlob(pattern string) error {
	var err error
	g.T, err = template.ParseGlob(pattern)
	return err
}



