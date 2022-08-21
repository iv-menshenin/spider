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
	MaxLevel
)

func (l Level) String() string {
	switch l {

	case LevelImportNone:
		return "ImportNone"

	case LevelImportMethod:
		return "ImportMethod"

	case LevelImportFunc:
		return "ImportFunc"

	case LevelImportStruct:
		return "ImportStruct"

	default:
		return "UnknownDependency"
	}
}
