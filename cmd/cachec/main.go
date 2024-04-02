package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/kkyr/fig"
	"github.com/maqdev/cachec/config"
	"github.com/maqdev/cachec/templates"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

const CacheCVersion = "0.0.1"
const DALStructName = "Queries"

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
	typeMap := mergeTypeMap(cfg.Types)
	convertFunctions := mergeConvertFunctionMap(cfg.ConvertFuncs)
	protoImports := mergeProtoImportMap(cfg.ProtoImports)
	skipDALMethids := mergeSkipDALMethods(cfg.SkipDALMethods)

	for _, p := range cfg.Packages {
		slog.Info("Processing", "package", p.Source)

		cfg := &packages.Config{
			Mode: packages.NeedTypesInfo | packages.NeedTypes | packages.NeedImports | packages.NeedName | packages.NeedSyntax,
		}

		pkgs, err := packages.Load(cfg, string(moduleJoinPath(rootModuleName, p.Source)))
		if err != nil {
			return fmt.Errorf("failed to load package '%s': %w", p.Source, err)
		}

		for _, pkg := range pkgs {
			if len(pkg.Errors) != 0 {
				return fmt.Errorf("failed to load package '%s': %w", p.Source, pkg.Errors[0])
			}
			packageName := pkg.Name

			slog.Debug("Walking package", "name", packageName)
			if p.DALOutput == "" {
				return fmt.Errorf("DALOutput is required for package '%s'", packageName)
			}

			var protoGoPackagePath config.GoModule
			if p.CacheModelsOutput != "" {
				protoGoPackagePath = moduleRelPath(p.CacheModelsOutput)
			} else {
				protoGoPackagePath = moduleRelPath(p.DALOutput) + "/cache"
			}

			td := templateData{
				ProtoPackageName:    packageName,
				SourceGoPackagePath: moduleJoinPath(rootModuleName, p.Source),
				GoPackageName:       packageName,
				DALGoPackagePath:    moduleJoinPath(rootModuleName, p.DALOutput),
				ProtoGoPackagePath:  protoGoPackagePath,
				CacheCVersion:       CacheCVersion,
				GoCachePackageName:  path.Base(string(protoGoPackagePath)),
				GoCacheImports:      make(map[importAlias]config.GoModule), // todo: reset per file?
			}

			v := &astVisitor{templateData: td,
				typeMap:          typeMap,
				protoImportsMap:  protoImports,
				skipDALMethids:   skipDALMethids,
				convertFunctions: convertFunctions,
				cacheEntities:    p.Entities,
			}
			v.walkPackage(pkg)

			if v.err != nil {
				return fmt.Errorf("package parsing failed '%s': %w", packageName, v.err)
			}

			tmpls := []struct {
				dest     string
				template *template.Template
			}{
				{
					dest:     path.Join(p.ProtoOutput, packageName+".proto"),
					template: templates.ProtoTemplate,
				},
				{
					dest:     path.Join(p.DALOutput, packageName+".go"),
					template: templates.DALTemplate,
				},

				{
					dest:     path.Join(string(protoGoPackagePath), packageName+".protoconv.go"),
					template: templates.ProtoConvTemplate,
				},
			}

			for _, tmpl := range tmpls {
				err = os.MkdirAll(path.Dir(tmpl.dest), 0755)
				if err != nil {
					return fmt.Errorf("failed to create ouput directory '%s': %w", p.ProtoOutput, err)
				}

				slog.Info("Generating", "file", tmpl.dest)
				var f *os.File
				f, err = os.Create(tmpl.dest)
				if err != nil {
					return fmt.Errorf("failed to create file '%s': %w", tmpl.dest, err)
				}
				err = tmpl.template.Execute(f, v.templateData)
				if err != nil {
					return fmt.Errorf("failed to generate file: '%s': %w", tmpl.dest, err)
				}
			}
		}
	}

	return nil
}

func mergeSkipDALMethods(methods map[string]bool) map[string]bool {
	res := map[string]bool{
		"WithTx": true,
	}
	for k, v := range methods {
		res[k] = v
	}
	return res
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
		"time": map[config.GoType]config.ProtoType{
			"Time": "google.protobuf.Timestamp",
		},
		"github.com/jackc/pgx/v5/pgtype": map[config.GoType]config.ProtoType{
			"Text": "google.protobuf.StringValue",
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

func mergeConvertFunctionMap(convertFuncs config.ConvertFunctionMap) config.ConvertFunctionMap {
	res := config.ConvertFunctionMap{
		"google.protobuf.Timestamp": {
			// todo: time converters
			ToProto:   "github.com/maqdev/cachec/util/protoutil.PGToWrappedString",
			FromProto: "github.com/maqdev/cachec/util/protoutil.WrappedStringToPG",
		},
		"google.protobuf.StringValue": {
			ToProto:   "github.com/maqdev/cachec/util/protoutil.PGToWrappedString",
			FromProto: "github.com/maqdev/cachec/util/protoutil.WrappedStringToPG",
		},
	}

	for protoType, funcs := range convertFuncs {
		res[protoType] = funcs
	}
	return res
}

type ProtoImportMap map[config.ProtoType]config.ProtoFile

func mergeProtoImportMap(protoImports map[config.ProtoFile][]config.ProtoType) ProtoImportMap {
	res := ProtoImportMap{
		"google.protobuf.Timestamp":   "google/protobuf/timestamp.proto",
		"google.protobuf.StringValue": "google/protobuf/wrappers.proto",
	}
	for protoFile, protoTypes := range protoImports {
		for _, protoType := range protoTypes {
			res[protoType] = protoFile
		}
	}
	return res
}

func moduleJoinPath(rootPath string, subdir string) config.GoModule {
	return config.GoModule(rootPath) + "/" + moduleRelPath(subdir)
}

func moduleRelPath(subdir string) config.GoModule {
	return config.GoModule(strings.TrimLeft(strings.TrimLeft(subdir, "./"), "."))
}

type templateMessageField struct {
	Type   config.ProtoType
	Name   string
	Number int
}

type templateMessage struct {
	Name   string
	Fields []templateMessageField
}

type templateParam struct {
	Name       string
	GoFullType string
}

type templateDALMethod struct {
	Name   string
	Args   []templateParam
	Result []templateParam
}

type templateCacheEntityField struct {
	Name      string
	FromProto string
	ToProto   string
}

type templateCacheEntity struct {
	Name              string
	EntityImportAlias importAlias
	Fields            []templateCacheEntityField
}

type templateData struct {
	ProtoPackageName    string
	SourceGoPackagePath config.GoModule
	DALGoPackagePath    config.GoModule
	ProtoGoPackagePath  config.GoModule
	GoPackageName       string
	CacheCVersion       string
	ProtoImports        []config.ProtoFile
	ProtoMessages       []templateMessage
	DALMethods          []templateDALMethod
	GoImports           map[importAlias]config.GoModule

	GoCachePackageName string
	GoCacheImports     map[importAlias]config.GoModule
	CacheEntities      []templateCacheEntity
}

type importAlias string

type astVisitor struct {
	templateData     templateData
	typeMap          config.TypeMap
	protoImportsMap  ProtoImportMap
	goFileImports    map[importAlias]config.GoModule
	skipDALMethids   map[string]bool
	cacheEntities    map[string]config.EntityConfig
	convertFunctions config.ConvertFunctionMap
	err              error
}

func (v *astVisitor) walkPackage(pkg *packages.Package) {
	v.templateData.GoImports = make(map[importAlias]config.GoModule)
	for _, file := range pkg.Syntax {
		ast.Walk(v, file)
	}
}

func (v *astVisitor) Visit(n ast.Node) ast.Visitor {
	// fmt.Printf("------> %v %T\n", n, n)

	var err error
	switch t := n.(type) {
	case *ast.File:
		// reset file imports (aliases)
		v.goFileImports = make(map[importAlias]config.GoModule)

	case *ast.GenDecl:
		if t.Tok == token.TYPE {
			err = v.visitGenDecl(t)
		}

	case *ast.ImportSpec:
		err = v.visitImportSpec(t)

	case *ast.FuncDecl:
		err = v.visitFuncDecl(t)
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
			if structType, okStructType := typeSpec.Type.(*ast.StructType); okStructType {
				if typeSpec.Name.Name == DALStructName {
					err := v.visitQueries(structType)
					if err != nil {
						return err
					}
				}

				msg := templateMessage{
					Name: typeSpec.Name.Name,
				}

				index := 10
				for _, field := range structType.Fields.List {
					if len(field.Names) > 0 { // skip unnamed fields (anonymous structs?)
						fieldName := field.Names[0].Name
						if !unicode.IsUpper([]rune(fieldName)[0]) {
							continue
						}

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

				if len(msg.Fields) > 0 {
					v.templateData.ProtoMessages = append(v.templateData.ProtoMessages, msg)

					if ce, ok := v.cacheEntities[msg.Name]; ok {
						v.templateData.CacheEntities = append(v.templateData.CacheEntities, v.createCacheEntity(msg, ce))
					}
				}
			}
		}
	}
	return nil
}

func (v *astVisitor) createCacheEntity(msg templateMessage, ce config.EntityConfig) templateCacheEntity {
	var fields []templateCacheEntityField
	for _, src := range msg.Fields {
		field := templateCacheEntityField{
			Name: src.Name,
		}
		if conv, ok := v.convertFunctions[src.Type]; ok {
			field.FromProto = addOrFindAliasFunc(conv.FromProto, v.templateData.GoCacheImports)
			field.ToProto = addOrFindAliasFunc(conv.ToProto, v.templateData.GoCacheImports)
			//v.goFileImports
		}

		fields = append(fields, field)
	}

	return templateCacheEntity{
		Name:              msg.Name,
		EntityImportAlias: addOrFindAlias(v.templateData.SourceGoPackagePath, v.templateData.GoCacheImports),
		Fields:            fields,
	}
}

func addOrFindAliasFunc(fnc string, imports map[importAlias]config.GoModule) string {
	sep := strings.LastIndex(fnc, ".")
	if sep < 0 {
		return fnc
	}
	mod := fnc[:sep]
	funcName := fnc[sep+1:]
	alias := addOrFindAlias(config.GoModule(mod), imports)
	return string(alias) + "." + funcName
}

func (v *astVisitor) visitImportSpec(t *ast.ImportSpec) error {
	var alias importAlias
	var mod config.GoModule
	if t.Name != nil {
		alias = importAlias(t.Name.Name)
	} else {
		mods, err := strconv.Unquote(t.Path.Value)
		if err != nil {
			return fmt.Errorf("failed to parse import path '%s' : %w", t.Path.Value, err)
		}
		mod = config.GoModule(mods)
		alias = defaultAlias(mod)
	}
	if v.goFileImports == nil {
		v.goFileImports = make(map[importAlias]config.GoModule)
	}
	v.goFileImports[alias] = mod
	return nil
}

var versionedPackageRegex = regexp.MustCompile("^v\\d+$")

func defaultAlias(p config.GoModule) importAlias {
	s := path.Base(string(p))
	if versionedPackageRegex.Match([]byte(s)) {
		s = path.Base(path.Dir(string(p)))
	}
	return importAlias(s)
}

func (v *astVisitor) resolveGoType(expr ast.Expr) (config.GoModule, config.GoType, error) {
	var module config.GoModule
	var typ config.GoType

	switch t := expr.(type) {
	case *ast.Ident:
		typ = config.GoType(t.Name)
	case *ast.SelectorExpr:
		alias := importAlias(t.X.(*ast.Ident).Name)
		if m, ok := v.goFileImports[alias]; ok {
			module = m
		} else {
			return "", "", fmt.Errorf("unknown import alias: %v", alias)
		}

		typ = config.GoType(t.Sel.Name)
	case *ast.ArrayType:
		var err error
		module, typ, err = v.resolveGoType(t.Elt)
		if err != nil {
			return "", "", err
		}
		typ = "[]" + typ
	default:
		return "", "", fmt.Errorf("unknown field type: %v", expr)
	}
	// source package types if they start with Uppercase, otherwise those are builtin types such as int32
	if module == "" && unicode.IsUpper([]rune(typ)[0]) {
		module = v.templateData.SourceGoPackagePath
	}

	return module, typ, nil
}

func (v *astVisitor) structTypeToProto(expr ast.Expr) (config.ProtoType, error) {
	module, typ, err := v.resolveGoType(expr)
	if err != nil {
		return "", err
	}

	var isArray bool
	if strings.HasPrefix(string(typ), "[]") {
		isArray = true
		typ = config.GoType(strings.TrimLeft(string(typ), "[]"))
	}

	if types, ok := v.typeMap[module]; ok {
		if protoType, okProtoType := types[typ]; okProtoType {
			if protoFile, okProtoImport := v.protoImportsMap[protoType]; okProtoImport {
				if slices.Index(v.templateData.ProtoImports, protoFile) < 0 {
					v.templateData.ProtoImports = append(v.templateData.ProtoImports, protoFile)
				}
			}

			// todo: apply isArray !!!
			fmt.Println(isArray)

			return protoType, nil
		}
	}
	return "", fmt.Errorf("can't find mapping of %s.%s", module, typ)
}

func (v *astVisitor) visitQueries(_ *ast.StructType) error {
	return nil
}

func (v *astVisitor) visitFuncDecl(t *ast.FuncDecl) error {
	if t.Name == nil ||
		v.skipDALMethids[t.Name.Name] ||
		!unicode.IsUpper([]rune(t.Name.Name)[0]) ||
		t.Type.Params == nil || len(t.Type.Params.List) == 0 ||
		t.Type.Results == nil || len(t.Type.Results.List) == 0 ||
		t.Recv == nil || len(t.Recv.List) != 1 {
		return nil
	}

	receiverName := resolveReceiver(t)

	// handle `Queries` methods to generate DAL interface
	if receiverName == DALStructName {
		args, err := v.resolveListOfArgs(t.Type.Params)
		if err != nil {
			return err
		}
		var result []templateParam
		result, err = v.resolveListOfArgs(t.Type.Results)
		if err != nil {
			return err
		}

		m := templateDALMethod{
			Name:   t.Name.Name,
			Args:   args,
			Result: result,
		}
		v.templateData.DALMethods = append(v.templateData.DALMethods, m)
	}
	return nil
}

func varName(name []*ast.Ident) string {
	if name != nil {
		return name[0].Name
	}
	return ""
}

func resolveReceiver(t *ast.FuncDecl) string {
	if t.Recv == nil || len(t.Recv.List) != 1 {
		return ""
	}

	switch recv := t.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return recv.Name
	case *ast.StarExpr:
		return recv.X.(*ast.Ident).Name
	}
	return ""
}

func (v *astVisitor) resolveListOfArgs(params *ast.FieldList) ([]templateParam, error) {
	res := make([]templateParam, 0, len(params.List))
	for _, p := range params.List {
		mod, typ, err := v.resolveGoType(p.Type)
		if err != nil {
			return nil, err
		}
		alias := addOrFindAlias(mod, v.templateData.GoImports)

		var fullType string
		if alias != "" {
			if strings.HasPrefix(string(typ), "[]") {
				typ = typ[2:]
				alias = "[]" + alias
			}

			fullType = string(alias) + "." + string(typ)
		} else {
			fullType = string(typ)
		}

		res = append(res, templateParam{
			Name:       varName(p.Names),
			GoFullType: fullType,
		})
	}
	return res, nil
}

func addOrFindAlias(mod config.GoModule, imports map[importAlias]config.GoModule) importAlias {
	if mod == "" {
		return ""
	}
	def := defaultAlias(mod)
	alias := def
	for i := 1; ; i++ {
		if existingMod, ok := imports[alias]; ok {
			if existingMod == mod {
				break
			}
		} else {
			imports[alias] = mod
			break
		}
		alias = importAlias(fmt.Sprintf("%s%d", def, i))
	}
	return alias
}

var emptyVisitor ast.Visitor = emptyVisitorImpl{}

type emptyVisitorImpl struct{}

func (e emptyVisitorImpl) Visit(_ ast.Node) (w ast.Visitor) { return e }
