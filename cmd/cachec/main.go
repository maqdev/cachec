package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/kkyr/fig"
	"github.com/maqdev/cachec/config"
	"github.com/maqdev/cachec/templates"
	"golang.org/x/mod/modfile"
)

const CacheCVersion = "0.0.1"

func main() {
	err := run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
}

func run() error {
	configPathFlag := flag.String("config", "cachec.yaml", "Path to the config file")
	envPrefix := flag.String("env-prefix", "CACHEC", "Prefix to use for environment variables")
	goModPath := flag.String("modfile", "go.mod", "Path to go.mod file")
	flag.Parse()
	configPath := *configPathFlag

	var cfg config.AppConfig
	dir := fig.Dirs(".")
	if path.IsAbs(configPath) {
		dir = fig.Dirs(path.Dir(configPath))
		configPath = path.Base(configPath)
	}
	err := fig.Load(&cfg, dir, fig.File(configPath), fig.UseEnv(*envPrefix))
	if err != nil {
		return fmt.Errorf("failed to load config file '%s': %w", *configPathFlag, err)
	}

	mfContent, err := os.ReadFile(*goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod file '%s': %w", *goModPath, err)
	}

	modFileParsed, err := modfile.ParseLax(*goModPath, mfContent, nil)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod file '%s': %w", *goModPath, err)
	}

	rootModuleName := modFileParsed.Module.Mod.Path
	typeMap := mergeTypeMap(cfg.TypeMap)

	for _, p := range cfg.Packages {
		slog.Info("Processing", "package", p.Source)

		fs := token.NewFileSet()
		f, err := parser.ParseDir(fs, "./gen/a/", nil, parser.AllErrors /* parser.SkipObjectResolution*/)
		if err != nil {
			return fmt.Errorf("failed to parse package '%s': %w", p.Source, err)
		}

		err = os.MkdirAll(p.ProtoOutput, 0755)
		if err != nil {
			return fmt.Errorf("failed to create ouput directory '%s': %w", p.ProtoOutput, err)
		}

		for packageName, packageAst := range f {
			slog.Debug("Walking package", "name", packageName)

			td := templateData{
				Package:       packageName,
				GoPackage:     moduleJoinPath(rootModuleName, p.ProtoOutput),
				CacheCVersion: CacheCVersion,
			}

			v := &astVisitor{templateData: td, typeMap: typeMap}
			ast.Walk(v, packageAst)

			if v.err != nil {
				return fmt.Errorf("failed to walk package '%s': %w", packageName, v.err)
			}

			f, err := os.Create(path.Join(p.ProtoOutput, packageName+".proto"))
			if err != nil {
				return fmt.Errorf("failed to create proto file '%s': %w", packageName, err)
			}
			err = templates.ProtoTemplate.Execute(f, v.templateData)
			if err != nil {
				return fmt.Errorf("failed to generate proto file: '%s': %w", packageName, err)
			}
		}
	}

	return nil
}

func mergeTypeMap(typeMap config.TypeMap) config.TypeMap {
	res := config.TypeMap{
		"": map[config.GoType]config.ProtoType{
			"int32":   "int32",
			"int64":   "int64",
			"uint32":  "uint32",
			"uint64":  "uint64",
			"float32": "float",
			"float64": "double",
			"bool":    "bool",
			"string":  "string",
			"[]byte":  "bytes",
		},
	}

	for goModule, types := range typeMap {
		if _, ok := res[goModule]; !ok {
			res[goModule] = make(map[config.GoType]config.ProtoType)
		}
		for goType, protoType := range types {
			res[goModule][goType] = protoType
		}
	}
	return res
}

func moduleJoinPath(name string, output string) string {
	return name + "/" + strings.TrimLeft(strings.TrimLeft(output, "./"), ".")
}

type templateMessageField struct {
	Type   string
	Name   string
	Number int
}

type templateMessage struct {
	Name   string
	Fields []templateMessageField
}

type templateData struct {
	Package       string
	GoPackage     string
	CacheCVersion string
	Messages      []templateMessage
}

type importAlias string

type astVisitor struct {
	imports      map[importAlias]config.GoModule
	templateData templateData
	typeMap      config.TypeMap
	err          error
}

func (v *astVisitor) Visit(n ast.Node) ast.Visitor {
	var err error
	switch t := n.(type) {
	case *ast.GenDecl:
		if t.Tok == token.TYPE {
			err = v.visitGenDecl(t)

		}
	case *ast.ImportSpec:
		err = v.visitImportSpec(t)
	}
	if err != nil {
		v.err = err
		return emptyVisitor
	}
	return v
}

func (v *astVisitor) visitGenDecl(t *ast.GenDecl) error {
	for _, spec := range t.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				msg := templateMessage{
					Name: typeSpec.Name.Name,
				}

				index := 10
				for _, field := range structType.Fields.List {
					if len(field.Names) > 0 { // skip unnamed fields (anonymous structs?)
						fieldName := field.Names[0].Name
						typ, err := v.structTypeToProto(field.Type)
						if err != nil {
							return fmt.Errorf("failed to convert field '%s' of type '%s' to proto: %w", fieldName, field.Type, err)
						}
						msg.Fields = append(msg.Fields, templateMessageField{
							Type:   typ,
							Name:   fieldName,
							Number: index,
						})
						index += 10
					}
				}

				v.templateData.Messages = append(v.templateData.Messages, msg)
			}
		}
	}
	return nil
}

func (v *astVisitor) visitImportSpec(t *ast.ImportSpec) error {
	var alias string
	if t.Name != nil {
		alias = t.Name.Name
	} else {
		alias = path.Base(t.Path.Value)
	}
	if v.imports == nil {
		v.imports = make(map[importAlias]config.GoModule)
	}
	v.imports[importAlias(alias)] = config.GoModule(t.Path.Value)
	return nil
}

func (v *astVisitor) structTypeToProto(expr ast.Expr) (string, error) {
	var module config.GoModule
	var typ config.GoType

	switch t := expr.(type) {
	case *ast.Ident:
		typ = config.GoType(t.Name)
	case *ast.SelectorExpr:
		alias := importAlias(t.X.(*ast.Ident).Name)
		if m, ok := v.imports[alias]; ok {
			module = m
		} else {
			return "", fmt.Errorf("unknown import alias: %v", alias)
		}

		typ = config.GoType(t.Sel.Name)
	default:
		return "", fmt.Errorf("Unknown field type: %v", expr)
	}

	if types, ok := v.typeMap[module]; ok {
		if protoType, ok := types[typ]; ok {
			return string(protoType), nil
		}
	}
	return "", fmt.Errorf("Can't find mapping of %s.%s", typ, module)
}

var emptyVisitor ast.Visitor = emptyVisitorImpl{}

type emptyVisitorImpl struct{}

func (e emptyVisitorImpl) Visit(node ast.Node) (w ast.Visitor) { return e }
