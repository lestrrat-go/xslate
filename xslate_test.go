package xslate

import (
  "testing"
  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/parser/tterse"
  "github.com/lestrrat/go-xslate/vm"
//  txtime "github.com/lestrrat/go-xslate/functions/time"
)

func ExampleXslate () {
/*
  tx := xslate.New()
  tx.RegisterFunctions(txtime.New())
  tx.RenderString(template)
*/
}

// TODO: vm.Vars should be xslate.Vars?
func executeAndCompare(t *testing.T, template string, vars vm.Vars, expected string) {
  p := tterse.New()
  c := compiler.New()
  ast, err := p.Parse(template)
  if err != nil {
    t.Fatalf("Failed to parse template: %s", err)
  }

  bc, err := c.Compile(ast)
  if err != nil {
    t.Fatalf("Failed to compile ast: %s", err)
  }

t.Logf("bytecode = %s\n", bc)

  v := vm.NewVM()
  v.Run(bc, vars)

  output, err := v.OutputString()
  if err != nil {
    t.Fatalf("Failed to get output from virtual machine: %s", err)
  }
  if output != expected {
    t.Errorf("Expected '%s', got '%s'", expected, output)
  }
}

func TestXslate_SimpleString(t *testing.T) {
  executeAndCompare(t, `Hello, World!`, nil, `Hello, World!`)
}

func TestXslate_LocalVar(t *testing.T) {
  executeAndCompare(t, `Hello World, [% name %]!`, vm.Vars { "name": "Bob" }, `Hello World, Bob!`)
}
