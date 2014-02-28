package tterse

import (
  "fmt"
//  "github.com/lestrrat/go-xslate/ast"
  "github.com/lestrrat/go-xslate/parser"
)

const (
  ItemWrapper parser.LexItemType = parser.DefaultItemTypeMax + 1
)

var operators = map[string]parser.LexItemType{
  "+":  parser.ItemPlus,
  "=":  parser.ItemAssign,
}

var symbols = map[string]parser.LexItemType{
  "WRAPPER":  ItemWrapper,
  "SET":      parser.ItemSet,
  "GET":      parser.ItemGet,
  "IF":       parser.ItemIf,
  "ELSIF":    parser.ItemElseIf,
  "ELSE":     parser.ItemElse,
  "UNLESS":   parser.ItemUnless,
  "FOREACH":  parser.ItemForeach,
  "IN":       parser.ItemIn,
  "MACRO":    parser.ItemMacro,
  "BLOCK":    parser.ItemBlock,
  "END":      parser.ItemEnd,
}

type AST struct {}
type Lexer struct {
  *parser.Lexer
}

type TTerse struct {
  lexer *Lexer
}

func NewLexer() *Lexer {
  l := &Lexer {
    parser.NewLexer(),
  }
  l.SetTagStart("[%")
  l.SetTagEnd("%]")
  for k, v := range symbols {
    l.AddSymbol(k, v)
  }
  for k, v := range operators {
    l.AddSymbol(k, v)
  }
  return l
}

func New() *TTerse {
  return &TTerse {
    lexer: NewLexer(),
  }
}

func (p *TTerse) Parse(input string) (*AST, error) {
  p.lexer.SetInput(input)
  go p.lexer.Run()

Loop:
  for {
    item := p.lexer.NextItem()
    switch item.Type() {
    case parser.ItemEnd:
      break Loop
    case parser.ItemRawString:
    default:
      fmt.Printf("item -> %v\n", item)
    }
  }

  return nil, nil
}