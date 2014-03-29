package kolonish

import (
  "testing"
  "github.com/lestrrat/go-lex"
  "github.com/lestrrat/go-xslate/parser"
)

func makeItem(t lex.ItemType, p, line int, v string) lex.LexItem {
  return lex.NewItem(t, p, line, v)
}

func makeLexer(input string) *parser.Lexer {
  l := NewLexer(input)
  return l
}

func lexit(input string) *parser.Lexer {
  l := makeLexer(input)
  go l.Run(l)
  return l
}

func compareLex(t *testing.T, expected []lex.LexItem, l *parser.Lexer) {
  for n := 0; n < len(expected); n++ {
    i := l.NextItem()

    e := expected[n]
    if e.Type() != i.Type() {
      t.Errorf("Expected type %s, got %s", e.Type(), i.Type())
      t.Logf("   -> expected %s got %s", e, i)
    }
    if e.Type() == parser.ItemIdentifier || e.Type() == parser.ItemRawString {
      if e.Value() != i.Value() {
        t.Errorf("Expected.Value()ue %s, got %s", e.Value(), i.Value())
        t.Logf("   -> expected %s got %s", e, i)
      }
    }
  }

  i := l.NextItem()
  if i.Type() != parser.ItemEOF {
    t.Errorf("Expected EOF, got %s", i)
  }

}

func TestGetImplicit(t *testing.T) {
  tmpl  := `<: $foo :>`
  l     := lexit(tmpl)
  expected := []lex.LexItem {
    makeItem(parser.ItemTagStart, 0, 1, "<:"),
    makeItem(parser.ItemSpace, 2, 1, " "),
    makeItem(ItemDollar, 3, 1, "$"),
    makeItem(parser.ItemIdentifier, 4, 1, "foo"),
    makeItem(parser.ItemSpace, 7, 1, " "),
    makeItem(parser.ItemTagEnd, 8, 1, ":>"),
  }
  compareLex(t, expected, l)
}

func TestLinewiseCode(t *testing.T) {
  tmpl := `
: "foo\n"
: for list -> i {
:    i
: }
`
  _ = lexit(tmpl)

}