package xslate

import (
  "fmt"

  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/loader"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/parser/tterse"
  "github.com/lestrrat/go-xslate/vm"
)

const (
  DUMP_BYTECODE = 1 << iota
  DUMP_AST
)

type Vars vm.Vars
type Xslate struct {
  Flags    int32
  Vm       *vm.VM
  Compiler compiler.Compiler
  Parser   parser.Parser
  Loader   loader.Loader
  // XXX Need to make syntax pluggable
}

func New() *Xslate {
  return &Xslate{
    Vm:       vm.NewVM(),
    Compiler: compiler.New(),
    Parser:   tterse.New(),
    Loader:   nil, // Loader is not necessary if you're just doing
                   // RenderString(). But to load files, you need to set
                   // this up somehow
  }
}

func (x *Xslate) Render(name string, vars Vars) (string, error) {
  template, err := x.Loader.Load(name)
  if err != nil {
    return "", err
  }

  return x.RenderString(string(template), vars)
}

func (x *Xslate) RenderString(template string, vars Vars) (string, error) {
  ast, err := x.Parser.ParseString(template)
  if err != nil {
    return "", err
  }

  if x.Flags & DUMP_AST != 0 {
    fmt.Printf("%s\n", ast)
  }

  bc, err := x.Compiler.Compile(ast)
  if err != nil {
    return "", err
  }

  if x.Flags & DUMP_BYTECODE != 0 {
    fmt.Printf("%s\n", bc)
  }

  x.Vm.Run(bc, vm.Vars(vars))
  str, err := x.Vm.OutputString()
  return str, err
}
