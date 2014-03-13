package xslate

import (
  "testing"
//  txtime "github.com/lestrrat/go-xslate/functions/time"
)

func ExampleXslate () {
/*
  tx := xslate.New()
  tx.RegisterFunctions(txtime.New())
  tx.RenderString(template)
*/
}

func executeAndCompare(t *testing.T, template string, vars Vars, expected string) {
  x := New()
  x.Flags |= DUMP_AST
  x.Flags |= DUMP_BYTECODE
  output, err := x.RenderString(template, vars)
  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  if output != expected {
    t.Errorf("Expected '%s', got '%s'", expected, output)
  }
}

func TestXslate_SimpleString(t *testing.T) {
  executeAndCompare(t, `Hello, World!`, nil, `Hello, World!`)
}

func TestXslate_Variable(t *testing.T) {
  executeAndCompare(t, `Hello World, [% name %]!`, Vars { "name": "Bob" }, `Hello World, Bob!`)
}

func TestXslate_LocalVar(t *testing.T) {
  executeAndCompare(t, `[% SET name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
}

/* I'm still very confused about the for_start / for_iter bytecodes 
func TestXslate_Foreach(t *testing.T) {
  var list [10]int
  for i := 0; i < 10; i++ {
    list[i] = i
  }
  template := `[% FOREACH i IN list %][% i %],[% END %]`
  executeAndCompare(t, template, Vars { "list": list }, `0, 1, 2, 3, 4, 5, 6, 7, 8, 9,`)
}
*/
