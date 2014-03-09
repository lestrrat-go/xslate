package tterse

import (
  "testing"
)

func TestBasic(t *testing.T) {
  tmpl := `
[% WRAPPER hoge.tx WITH foo = "bar" %]
[% FOREACH x IN list %]
[% loop.index %]. x is [% x %]
[% END %]
[% END %]
`
  p   := New()
  ast, err := p.Parse(tmpl)
  t.Logf("ast = %s, err = %s", ast, err)
}

func TestSimpleAssign(t *testing.T) {
  tmpl := `[% s = 1 %]`
  p := New()
  ast, _ := p.Parse(tmpl)
  t.Logf("%#v", ast)
}
