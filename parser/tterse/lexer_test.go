package tterse

import (
  "testing"
  "github.com/lestrrat/go-xslate/parser"
)

func makeItem(t parser.LexItemType, p int, v string) parser.LexItem {
  return parser.NewLexItem(t, p, v)
}

var space     parser.LexItem = makeItem(parser.ItemSpace, 0, " ")
var tagStart  parser.LexItem = makeItem(parser.ItemTagStart, 0, "[%")
var tagEnd    parser.LexItem = makeItem(parser.ItemTagEnd, 0, "[%")
func makeLexer(input string) *Lexer {
  l := NewLexer()
  l.SetInput(input)
  return l
}

func lexit(input string) *Lexer {
  l := makeLexer(input)
  go l.Run()
  return l
}

func compareLex(t *testing.T, expected []parser.LexItem, l *Lexer) {
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

func TestLexRawString(t *testing.T) {
  tmpl  := `This is a raw string 日本語もはいるよ！`
  l     := lexit(tmpl)

  for {
    i := l.NextItem()
    if i.Type() == parser.ItemEOF || i.Type() == parser.ItemError {
      break
    }

    if i.Type() != parser.ItemRawString {
      t.Errorf("Expected type RawString, got %s", i)
    }

    if i.Value() != tmpl {
      t.Errorf("Expected.Value()ue '%s', got '%s'", tmpl, i.Value())
    }
  }
}

func TestLexSet(t *testing.T) {
  tmpl  := `[% SET foo = bar + 1 %]`
  l     := lexit(tmpl)
  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemSet,        0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    space,
    makeItem(parser.ItemAssign,     0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "bar"),
    space,
    makeItem(parser.ItemPlus,       0, ""),
    space,
    makeItem(parser.ItemNumber,     0, "1"),
    space,
    tagEnd,
  }
  compareLex(t, expected, l)
}

func TestLexGet(t *testing.T) {
  tmpl  := `[% GET foo %]`
  l     := lexit(tmpl)
  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemGet,        0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    space,
    tagEnd,
  }
  compareLex(t, expected, l)
}

func TestLexForeach(t *testing.T) {
  tmpl  := `[% FOREACH i IN list %][% i %][% END %]`
  l     := lexit(tmpl)
  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemForeach,    0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "i"),
    space,
    makeItem(parser.ItemIn,         0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "list"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemIdentifier, 0, "i"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemEnd,        0, ""),
    space,
    tagEnd,
  }

  compareLex(t, expected, l)
}

func TestLexMacro(t *testing.T) {
  tmpl  := `[% MACRO foo BLOCK %]foo bar[% baz %][% END %]`
  l     := lexit(tmpl)
  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemMacro,    0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    space,
    makeItem(parser.ItemBlock,      0, ""),
    space,
    tagEnd,
    makeItem(parser.ItemRawString,  0, "foo bar"),
    tagStart,
    space,
    makeItem(parser.ItemIdentifier, 0, "baz"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemEnd,        0, ""),
    space,
    tagEnd,
  }

  compareLex(t, expected, l)
}

func TestLexConditional(t *testing.T) {
  tmpl  := `[% IF foo %][% IF (bar) %]baz[% END %][% ELSIF quux %]hoge[% ELSE %]fuga[% END %][% UNLESS moge %]bababa[% END %]`
  l     := lexit(tmpl)
  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemIf,         0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemIf,         0, ""),
    space,
    makeItem(parser.ItemOpenParen,  0, ""),
    makeItem(parser.ItemIdentifier, 0, "bar"),
    makeItem(parser.ItemCloseParen, 0, ""),
    space,
    tagEnd,
    makeItem(parser.ItemRawString,  0, "baz"),
    tagStart,
    space,
    makeItem(parser.ItemEnd,        0, ""),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemElseIf,     0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "quux"),
    space,
    tagEnd,
    makeItem(parser.ItemRawString,  0, "hoge"),
    tagStart,
    space,
    makeItem(parser.ItemElse,       0, ""),
    space,
    tagEnd,
    makeItem(parser.ItemRawString,  0, "fuga"),
    tagStart,
    space,
    makeItem(parser.ItemEnd,        0, ""),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemUnless,     0, ""),
    space,
    makeItem(parser.ItemIdentifier, 0, "moge"),
    space,
    tagEnd,
    makeItem(parser.ItemRawString,  0, "bababa"),
    tagStart,
    space,
    makeItem(parser.ItemEnd,        0, ""),
    space,
    tagEnd,
  }
  compareLex(t, expected, l)
}

func TestVariableAccess(t *testing.T) {
  tmpl  := `[% foo.bar %][% foo.bar.baz() %]`
  l     := lexit(tmpl)

  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    makeItem(parser.ItemPeriod,     0, ""),
    makeItem(parser.ItemIdentifier, 0, "bar"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemIdentifier, 0, "foo"),
    makeItem(parser.ItemPeriod,     0, ""),
    makeItem(parser.ItemIdentifier, 0, "bar"),
    makeItem(parser.ItemPeriod,     0, ""),
    makeItem(parser.ItemIdentifier, 0, "baz"),
    makeItem(parser.ItemOpenParen,  0, ""),
    makeItem(parser.ItemCloseParen, 0, ""),
    space,
    tagEnd,
  }

  compareLex(t, expected, l)
}

func TestBareQuotedString(t *testing.T) {
  tmpl  := `[% "hello, double quote" %][% 'hello, single quote' %]`
  l     := lexit(tmpl)

  expected := []parser.LexItem {
    tagStart,
    space,
    makeItem(parser.ItemDoubleQuotedString, 0, "hello, double quote"),
    space,
    tagEnd,
    tagStart,
    space,
    makeItem(parser.ItemSingleQuotedString, 0, "hello, single quote"),
    space,
    tagEnd,
  }
  compareLex(t, expected, l)
}

