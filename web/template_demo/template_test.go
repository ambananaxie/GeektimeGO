package template_demo

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	type User struct {
		Name string
	}
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{ .Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Name: "Tom"})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())
}

func TestMapData(t *testing.T) {

	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{ .Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, map[string]string{"Name": "Tom"})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())
}

func TestSlice(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{index . 0}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, []string{"Tom"})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())
}

func TestBasic(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{.}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, 123)
	require.NoError(t, err)
	assert.Equal(t, `Hello, 123`, buffer.String())
}

func TestFuncCall(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
切片长度: {{len .Slice}}
{{printf "%.2f" 1.2345}}
Hello, {{.Hello "Tom" "Jerry"}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, `
切片长度: 2
1.23
Hello, Tom·Jerry`, buffer.String())
}

func TestLoop(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- range $idx, $ele := .Slice}}
{{- .}}
{{$idx}}-{{$ele}}
{{end}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, `a
0-a
b
1-b

`, buffer.String())
}

func TestForLoop(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- range $idx, $ele := .}}
{{- $idx}},
{{- end}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, make([]int, 100))
	require.NoError(t, err)
	assert.Equal(t, `0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,
`, buffer.String())
}

func TestIfElse(t *testing.T) {
	type User struct {
		Age int
	}
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- if and (gt .Age 0) (le .Age 6)}}
我是儿童: (0, 6]
{{ else if and (gt .Age 6) (le .Age 18) }}
我是少年: (6, 18]
{{ else }}
我是成人: >18
{{end -}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Age: 19})
	require.NoError(t, err)
	assert.Equal(t, `
我是成人: >18
`, buffer.String())
}

func TestPipeline(t *testing.T) {
	testCases := []struct{
		name string

		tpl  string
		data any

		want string
	} {
		// 这些例子来自官方文档
		// https://pkg.go.dev/text/template#hdr-Pipelines
		{
			name: "string constant",
			tpl:`{{"\"output\""}}`,
			want: `"output"`,
		},
		{
			name: "raw string constant",
			tpl: "{{`\"output\"`}}",
			want: `"output"`,
		},
		{
			name: "function call",
			tpl: `{{printf "%q" "output"}}`,
			want: `"output"`,
		},
		{
			name: "take argument from pipeline",
			tpl: `{{"output" | printf "%q"}}`,
			want: `"output"`,
		},
		{
			name: "parenthesized argument",
			tpl: `{{printf "%q" (print "out" "put")}}`,
			want: `"output"`,
		},
		{
			name: "elaborate call",
			// printf "%s%s" "out" "put"
			tpl: `{{"put" | printf "%s%s" "out" | printf "%q"}}`,
			want: `"output"`,
		},
		{
			name: "longer chain",
			tpl: `{{"output" | printf "%s" | printf "%q"}}`,
			want: `"output"`,
		},
		{
			name: "with action using dot",
			tpl: `{{with "output"}}{{printf "%q" .}}{{end}}`,
			want: `"output"`,
		},
		{
			name: "with action that creates and uses a variable",
			tpl: `{{with $x := "output" | printf "%q"}}{{$x}}{{end}}`,
			want: `"output"`,
		},
		{
			name: "with action that uses the variable in another action",
			tpl: `{{with $x := "output"}}{{printf "%q" $x}}{{end}}`,
			want: `"output"`,
		},
		{
			name: "pipeline with action that uses the variable in another action",
			tpl: `{{with $x := "output"}}{{$x | printf "%q"}}{{end}}`,
			want: `"output"`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tpl := template.New(tc.name)
			tpl, err := tpl.Parse(tc.tpl)
			if err != nil {
				t.Fatal(err)
			}
			bs := &bytes.Buffer{}
			err = tpl.Execute(bs, tc.data)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.want, bs.String())
		})
	}
}

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first string, last string) string {
	return fmt.Sprintf("%s·%s",  first, last)
}