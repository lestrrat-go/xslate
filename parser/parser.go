package parser

type Parser interface {
  Parse(string) (*AST, error)
}

