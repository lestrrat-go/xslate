package kolonish

import (
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-lex"
)

const (
  ItemDollar lex.LexItemType = parser.DefaultItemTypeMax + 1
)

var SymbolSet = parser.DefaultSymbolSet.Copy()
func init() {
  SymbolSet.Set("$", ItemDollar)
}

func NewLexer(template string) *parser.Lexer {
  l := parser.NewLexer(template, SymbolSet)
  l.SetTagStart("<:")
  l.SetTagEnd(":>")

  return l
}