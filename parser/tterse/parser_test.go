package tterse

import (
  "testing"
)

func TestBasic(t *testing.T) {
  tmpl := `
[% WRAPPER "hoge.tx" WITH foo = "bar" %]
[% FOREACH x IN list %]
[% loop.index %]. x is [% x %]
[% END %]
[% END %]
`
  p   := New()
  ast, err := p.Parse(tmpl)
  if err != nil {
    t.Errorf("Error during parse: %s", err)
  }

  if len(ast.Root.Nodes) == 1 {
    t.Errorf("Expected Root node to have 1 child, got %d", len(ast.Root.Nodes))
  }
}

func TestSimpleAssign(t *testing.T) {
  tmpl := `[% s = 1 %]`
  p := New()
  ast, _ := p.Parse(tmpl)
  t.Logf("%#v", ast)
}
