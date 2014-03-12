package xslate

import(
  "io"
  "io/ioutil"

  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/parser/tterse"
  "github.com/lestrrat/go-xslate/vm"
)

type Xslate struct {
  Vm        *vm.VM
  Compiler  compiler.Compiler
  Parser    parser.Parser
  // XXX Need to make syntax pluggable
}

func New() *Xslate {
  return &Xslate {
    Vm: vm.NewVM(),
    Compiler: compiler.New(),
    Parser: tterse.New(),
  }
}

func (x *Xslate) RenderReader(rdr io.Reader, vars vm.Vars) (string, error) {
  tmpl, err := ioutil.ReadAll(rdr)
  if err != nil {
    return "", err
  }

  return x.Render(tmpl, vars)
}

func (x *Xslate) RenderString(template string, vars vm.Vars) (string, error) {
  return x.Render([]byte(template), vars)
}

func (x *Xslate) Render(template []byte, vars vm.Vars) (string, error) {
  ast, err := x.Parser.Parse(template)
  if err != nil {
    return "", err
  }

  bc, err := x.Compiler.Compile(ast)
  if err != nil {
    return "", err
  }

  x.Vm.Run(bc, vars)
  str, err := x.Vm.OutputString()
  return str, err
}
