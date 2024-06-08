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
	Source         string
	ProtoDir       string
	DALDir         string
	CacheModelsDir string
	CacheDir       string
	Entities       map[string]EntityConfig
}

type EntityConfig struct {
	Keys          []string
	PartitionKeys []string
	TTL           time.Duration
	Read          map[MethodName]ReadMethodConfig
	Invalidate    map[MethodName]InvalidateMethodConfig
}

type MethodName string

type ReadMethodConfig struct{}
type InvalidateMethodConfig struct{}

type GoModule string
type GoType string
type ProtoType string
type ProtoFile string

type TypeInfo struct {
	ProtoType     ProtoType
	ToProtoFunc   string
	FromProtoFunc string
}

type TypeMap map[GoModule]map[GoType]TypeInfo

type ConvertFunctions struct {
	ToProto   string
	FromProto string
}
type ConvertFunctionMap map[ProtoType]ConvertFunctions
