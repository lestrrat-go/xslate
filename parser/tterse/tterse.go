package tterse

import (
  "github.com/lestrrat/go-xslate/parser"
)

var operators = map[string]parser.LexItemType{
  "+":  parser.ItemPlus,
  "-":  parser.ItemMinus,
  "*":  parser.ItemAsterisk,
  "/":  parser.ItemSlash,
  "=":  parser.ItemAssign,
}

var symbols = map[string]parser.LexItemType{
  "WRAPPER":  parser.ItemWrapper,
  "SET":      parser.ItemSet,
  "GET":      parser.ItemGet,
  "IF":       parser.ItemIf,
  "ELSIF":    parser.ItemElseIf,
  "ELSE":     parser.ItemElse,
  "UNLESS":   parser.ItemUnless,
  "FOREACH":  parser.ItemForeach,
  "MACRO":    parser.ItemMacro,
  "BLOCK":    parser.ItemBlock,
  "END":      parser.ItemEnd,
}

// Lexer lexes tempaltes in TTerse syntax
type Lexer struct {
  *parser.Lexer
}

// TTerse is the main parser for TTerse
type TTerse struct {
  lexer *Lexer
  items []parser.LexItem
}

// NewLexer creates a new lexer
func NewLexer() *Lexer {
  l := &Lexer {
    parser.NewLexer(),
  }
  l.SetTagStart("[%")
  l.SetTagEnd("%]")

  // XXX TTerse specific
  l.AddSymbol("WITH", parser.ItemWith)
  l.AddSymbol("INCLUDE", parser.ItemInclude, 1.5)
  l.AddSymbol("IN", parser.ItemIn, 2.0)
  l.AddSymbol("END", parser.ItemEnd)

  for k, v := range symbols {
    l.AddSymbol(k, v)
  }
  for k, v := range operators {
    l.AddSymbol(k, v)
  }
  return l
}

// New creates a new TTerse parser
func New() *TTerse {
  return &TTerse {
    lexer: NewLexer(),
  }
}

// Parse parses the given template and creates an AST
func (p *TTerse) Parse(template []byte) (*parser.AST, error) {
  return p.ParseString(string(template))
}

// ParseString is the same as Parse, but receives a string instead of []byte
func (p *TTerse) ParseString(template string) (*parser.AST, error) {
  b := parser.NewBuilder()
  lex := NewLexer()
  lex.SetInput(template)
  return b.Parse("foo", template, lex)
}
