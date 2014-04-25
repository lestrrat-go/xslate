package parser

import (
	"io"
)

// Parser defines the interface for Xslate parsers
type Parser interface {
	Parse(string, []byte) (*AST, error)
	ParseString(string, string) (*AST, error)
	ParseReader(string, io.Reader) (*AST, error)
}
