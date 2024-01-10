package config

type AppConfig struct {
	Packages       []Package
	TypeMap        TypeMap
	ProtoImports   map[ProtoFile][]ProtoType
	SkipDALMethods map[string]bool
}

type Package struct {
	Source      string
	ProtoOutput string
	GoOutput    string
}

type GoModule string
type GoType string
type ProtoType string
type ProtoFile string
type TypeMap map[GoModule]map[GoType]ProtoType
