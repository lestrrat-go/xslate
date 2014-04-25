package kolonish

import (
	"github.com/lestrrat/go-lex"
	"github.com/lestrrat/go-xslate/parser"
	"io"
)

const (
	ItemDollar lex.ItemType = parser.DefaultItemTypeMax + 1
)

var SymbolSet = parser.DefaultSymbolSet.Copy()

func init() {
	SymbolSet.Set("$", ItemDollar)
}

// Kolonish is the main parser for Kolonish
type Kolonish struct{}

// NewStringLexer creates a new lexer
func NewStringLexer(template string) *parser.Lexer {
	l := parser.NewStringLexer(template, SymbolSet)
	l.SetTagStart("<:")
	l.SetTagEnd(":>")

	return l
}

// NewReaderLexer creates a new lexer
func NewReaderLexer(rdr io.Reader) *parser.Lexer {
	l := parser.NewReaderLexer(rdr, SymbolSet)
	l.SetTagStart("<:")
	l.SetTagEnd(":>")

	return l
}

// New creates a new Kolonish parser
func New() *Kolonish {
	return &Kolonish{}
}

func NewLexer(template string) *parser.Lexer {
	l := parser.NewStringLexer(template, SymbolSet)
	l.SetTagStart("<:")
	l.SetTagEnd(":>")

	return l
}

// Parse parses the given template and creates an AST
func (p *Kolonish) Parse(name string, template []byte) (*parser.AST, error) {
	return p.ParseString(name, string(template))
}

// ParseString is the same as Parse, but receives a string instead of []byte
func (p *Kolonish) ParseString(name, template string) (*parser.AST, error) {
	b := parser.NewBuilder()
	lex := NewStringLexer(template)
	return b.Parse(name, lex)
}

// ParseReader gets the template content from an io.Reader type
func (p *Kolonish) ParseReader(name string, rdr io.Reader) (*parser.AST, error) {
	b := parser.NewBuilder()
	lex := NewReaderLexer(rdr)
	return b.Parse(name, lex)
}
