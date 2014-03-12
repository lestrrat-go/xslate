package compiler

import (
  "github.com/lestrrat/go-xslate/parser/tterse"
  "testing"
)

func TestCompiler(t *testing.T) {
  c := New()
  if c == nil {
    t.Fatalf("Failed to instanticate compiler")
  }
}

func TestCompile_RawText(t *testing.T) {
  p := tterse.New()
  ast, err := p.Parse(`Hello, World!`)
  if err != nil {
    t.Fatalf("Failed to parse template: %s", err)
  }

  c := New()
  bc, err := c.Compile(ast)
  if err != nil {
    t.Fatalf("Failed to compile ast: %s", err)
  }

  t.Logf("-> %+v", bc)
}