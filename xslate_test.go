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

func TestCompile(t *testing.T) {
  p := tterse.New()
  c := compiler.New()
  ast, err := p.Parse(`Hello, World!`)
  if err != nil {
    t.Fatalf("Failed to parse template: %s", err)
  }

  bc, err := c.Compile(ast)
  if err != nil {
    t.Fatalf("Failed to compile ast: %s", err)
  }

  v := vm.NewVM()
  v.Run(bc)

  output, err := v.OutputString()
  if err != nil {
    t.Fatalf("Failed to get output from virtual machine: %s", err)
  }
  if output != `Hello, World!` {
    t.Errorf("Expected '%s', got '%s'", `Hello, World!`, output)
  }
}
