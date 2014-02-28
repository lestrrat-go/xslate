package kolonish

import (
  "github.com/lestrrat/go-xslate/parser"
)

const (
  ItemDollar parser.LexItemType = parser.DefaultItemTypeMax + 1
)

var operators = map[string]parser.LexItemType{
  "+":  parser.ItemPlus,
  "=":  parser.ItemAssign,
}

var symbols = map[string]parser.LexItemType{
  "$":  ItemDollar,
}

type Lexer struct {
  *parser.Lexer
}

func NewLexer() *Lexer {
  l := &Lexer {
    parser.NewLexer(),
  }
  l.SetTagStart("<:")
  l.SetTagEnd(":>")
  for k, v := range symbols {
    l.AddSymbol(k, v)
  }
  for k, v := range operators {
    l.AddSymbol(k, v)
  }

  return l
}