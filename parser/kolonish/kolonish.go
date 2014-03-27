package kolonish

import (
  "github.com/lestrrat/go-xslate/parser"
)

const (
  ItemDollar parser.LexItemType = parser.DefaultItemTypeMax + 1
)

var SymbolSet = parser.DefaultSymbolSet.Copy()
func init() {
  SymbolSet.Set("$", ItemDollar)
}

type Lexer struct {
  *parser.Lexer
}

func NewLexer() *Lexer {
  l := &Lexer {
    parser.NewLexer(SymbolSet),
  }
  l.SetTagStart("<:")
  l.SetTagEnd(":>")

  return l
}