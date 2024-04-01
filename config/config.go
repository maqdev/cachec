package config

import "time"

type AppConfig struct {
	Packages       []Package
	Types          TypeMap
	ConvertFuncs   ConvertFunctionMap
	ProtoImports   map[ProtoFile][]ProtoType
	SkipDALMethods map[string]bool
}

type Package struct {
	Source            string
	ProtoOutput       string
	DALOutput         string
	CacheModelsOutput string
	Entities          map[string]EntityConfig
}

type EntityConfig struct {
	Key        []string
	TTL        time.Duration
	Read       map[MethodName]ReadMethodConfig
	Invalidate map[MethodName]InvalidateMethodConfig
}

type MethodName string

type ReadMethodConfig struct{}
type InvalidateMethodConfig struct{}

type GoModule string
type GoType string
type ProtoType string
type ProtoFile string
type TypeMap map[GoModule]map[GoType]ProtoType

type ConvertFunctions struct {
	ToProto   string
	FromProto string
}
type ConvertFunctionMap map[ProtoType]ConvertFunctions
