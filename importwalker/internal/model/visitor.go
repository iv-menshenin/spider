package model

type (
	Imported interface {
		GetField(string) Imported
		GetResult(int) Imported
		GetLevel() []Level
		GetPackage() string
	}
	Level           int
	PackageLookuper interface {
		Lookup(packageName string) Parsed
	}
)

const (
	LevelImportNone Level = iota
	LevelImportMethod
	LevelImportFunc
	LevelImportStruct
)
