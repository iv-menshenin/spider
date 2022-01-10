package model

import "go/ast"

type (
	Parsed interface {
		GetPath() string
		Packages() []string
		Package(name string) *ast.Package
	}
	AST interface {
		Len() int
		Iter() <-chan Parsed
		Lookup(packageName string) Parsed
	}
)
