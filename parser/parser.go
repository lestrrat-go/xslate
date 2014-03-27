package parser

// Parser defines the interface for Xslate parsers
type Parser interface {
  Parse([]byte) (*AST, error)
  ParseString(string) (*AST, error)
}

