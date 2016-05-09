package compiler

import (
	"github.com/lestrrat/go-xslate/parser"
	"github.com/lestrrat/go-xslate/vm"
)

// Compiler is the interface to objects that can convert AST trees to
// actual Xslate Virtual Machine bytecode (see vm.ByteCode)
type Compiler interface {
	Compile(*parser.AST) (*vm.ByteCode, error)
}

type context struct {
	ByteCode *vm.ByteCode
}

// BasicCompiler is the default compiler used by Xslate
type BasicCompiler struct{}

// Optimizer is the interface of things that can optimize the ByteCode
type Optimizer interface {
	Optimize(*vm.ByteCode) error
}

// NaiveOptimizer is the default ByteCode optimizer
type NaiveOptimizer struct{}
