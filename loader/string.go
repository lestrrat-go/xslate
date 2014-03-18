package loader

import (
  "fmt"
  "os"
  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

type StringByteCodeLoader struct {
  *Flags
  Parser    parser.Parser
  Compiler  compiler.Compiler
}

func NewStringByteCodeLoader (p parser.Parser, c compiler.Compiler) *StringByteCodeLoader {
  return &StringByteCodeLoader { NewFlags(), p, c }
}

func (l *StringByteCodeLoader) LoadString(template string) (*vm.ByteCode, error) {
  ast, err := l.Parser.ParseString(template)
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

