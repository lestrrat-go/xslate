package parser

type Parser interface {
  Parse([]byte) (*AST, error)
  ParseString(string) (*AST, error)
}

