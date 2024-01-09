package config

type AppConfig struct {
	Module       string
	Packages     []Package
	TypeMap      TypeMap
	ProtoImports map[ProtoFile][]ProtoType
}

type Package struct {
	Source      string
	ProtoOutput string
}

type GoModule string
type GoType string
type ProtoType string
type ProtoFile string
type TypeMap map[GoModule]map[GoType]ProtoType
