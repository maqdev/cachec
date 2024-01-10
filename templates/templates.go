package templates

import (
	"embed"
	"text/template"
)

var (
	//go:embed proto.tmpl
	protoFS embed.FS

	ProtoTemplate = mustParse("proto.tmpl", protoFS)

	//go:embed dal.tmpl
	dalFS embed.FS

	DALTemplate = mustParse("dal.tmpl", dalFS)
)

func mustParse(name string, fs embed.FS) *template.Template {
	b, err := fs.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return template.Must(template.New(name).Parse(string(b)))
}
