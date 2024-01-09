package config

type AppConfig struct {
	Module   string
	Packages []Package
	TypeMap  TypeMap
}

type Package struct {
	Source      string
	ProtoOutput string
}

type GoModule string
type GoType string
type ProtoType string
type TypeMap map[GoModule]map[GoType]ProtoType
