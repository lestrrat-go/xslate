package tterse

import (
  "io"
  "github.com/lestrrat/go-lex"
  "github.com/lestrrat/go-xslate/parser"
)

// SymbolSet contains TTerse specific symbols
var SymbolSet = parser.DefaultSymbolSet.Copy()
func init() {
 // "In" must come before Include
  SymbolSet.Set("INCLUDE",  parser.ItemInclude, 2.0)
  SymbolSet.Set("IN",       parser.ItemIn,      1.5)
  SymbolSet.Set("WITH",     parser.ItemWith)
  SymbolSet.Set("CALL",     parser.ItemCall)
  SymbolSet.Set("END",      parser.ItemEnd)
  SymbolSet.Set("WRAPPER",  parser.ItemWrapper)
  SymbolSet.Set("SET",      parser.ItemSet)
  SymbolSet.Set("GET",      parser.ItemGet)
  SymbolSet.Set("IF",       parser.ItemIf)
  SymbolSet.Set("ELSIF",    parser.ItemElseIf)
  SymbolSet.Set("ELSE",     parser.ItemElse)
  SymbolSet.Set("UNLESS",   parser.ItemUnless)
  SymbolSet.Set("FOREACH",  parser.ItemForeach)
  SymbolSet.Set("WHILE",    parser.ItemWhile)
  SymbolSet.Set("MACRO",    parser.ItemMacro)
  SymbolSet.Set("BLOCK",    parser.ItemBlock)
  SymbolSet.Set("END",      parser.ItemEnd)
}

// TTerse is the main parser for TTerse
type TTerse struct {
  items []lex.LexItem
}

// NewStringLexer creates a new lexer
func NewStringLexer(template string) *parser.Lexer {
  l := parser.NewStringLexer(template, SymbolSet)
  l.SetTagStart("[%")
  l.SetTagEnd("%]")

  return l
}

// NewReaderLexer creates a new lexer
func NewReaderLexer(rdr io.Reader) *parser.Lexer {
  l := parser.NewReaderLexer(rdr, SymbolSet)
  l.SetTagStart("[%")
  l.SetTagEnd("%]")

  return l
}

// New creates a new TTerse parser
func New() *TTerse {
  return &TTerse {}
}

// Parse parses the given template and creates an AST
func (p *TTerse) Parse(name string, template []byte) (*parser.AST, error) {
  return p.ParseString(name, string(template))
}

// ParseString is the same as Parse, but receives a string instead of []byte
func (p *TTerse) ParseString(name, template string) (*parser.AST, error) {
  b := parser.NewBuilder()
  lex := NewStringLexer(template)
  return b.Parse(name, lex)
}

// ParseReader gets the template content from an io.Reader type
func (p *TTerse) ParseReader(name string, rdr io.Reader) (*parser.AST, error) {
  b := parser.NewBuilder()
  lex := NewReaderLexer(rdr)
  return b.Parse(name, lex)
}
