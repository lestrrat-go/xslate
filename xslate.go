package xslate

import (
  "fmt"
  "io"
  "io/ioutil"
  "os"

  "github.com/lestrrat/go-xslate/compiler"
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
  // XXX Need to make syntax pluggable
}

func New() *Xslate {
  return &Xslate{
    Vm:       vm.NewVM(),
    Compiler: compiler.New(),
    Parser:   tterse.New(),
  }
}

func (x *Xslate) RenderReader(rdr io.Reader, vars Vars) (string, error) {
  tmpl, err := ioutil.ReadAll(rdr)
  if err != nil {
    return "", err
  }
  return x.Render(tmpl, vars)
}

func (x *Xslate) RenderFile(filename string, vars Vars) (string, error) {
  file, err := os.Open(filename)
  if err != nil {
    return "", err
  }

  return x.RenderReader(file, vars)
}

func (x *Xslate) RenderString(template string, vars Vars) (string, error) {
  return x.Render([]byte(template), vars)
}

func (x *Xslate) Render(template []byte, vars Vars) (string, error) {
  ast, err := x.Parser.Parse(template)
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
