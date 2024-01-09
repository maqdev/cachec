package templates

import (
	"embed"
	"text/template"
)

var (
	//go:embed proto.tmpl
	protoFS embed.FS

	ProtoTemplate = MustParse("proto", protoFS)
)

func MustParse(name string, fs embed.FS) *template.Template {
	b, err := protoFS.ReadFile("proto.tmpl")
	if err != nil {
		panic(err)
	}

	return template.Must(template.New(name).Parse(string(b)))
}
