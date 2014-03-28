package loader

import (
  "fmt"
  "os"
  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

// StringByteCodeLoader is a fancy name for objects that can "given a template 
// string, parse and compile it". This is one of the most common operations
// that users want to do, but it needs to be separate from other loaders
// because there's no sane way to cache intermediate results, and therefore
// has significant performance penalty
type StringByteCodeLoader struct {
  *Flags
  Parser    parser.Parser
  Compiler  compiler.Compiler
}

// NewStringByteCodeLoader creates a new object
func NewStringByteCodeLoader (p parser.Parser, c compiler.Compiler) *StringByteCodeLoader {
  return &StringByteCodeLoader { NewFlags(), p, c }
}

// LoadString takes a template string and compiles it into vm.ByteCode
func (l *StringByteCodeLoader) LoadString(name string, template string) (*vm.ByteCode, error) {
  ast, err := l.Parser.ParseString(name, template)
  if err != nil {
    return nil, err
  }

  if l.ShouldDumpAST() {
    fmt.Fprintf(os.Stderr, "AST:\n%s\n", ast)
  }

  bc, err := l.Compiler.Compile(ast)
  if err != nil {
    return nil, err
  }

  if l.ShouldDumpByteCode() {
    fmt.Fprintf(os.Stderr, "ByteCode:\n%s\n", bc)
  }

  return bc, nil
}

