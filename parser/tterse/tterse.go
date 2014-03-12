package tterse

import (
  "github.com/lestrrat/go-xslate/parser"
)

var operators = map[string]parser.LexItemType{
  "+":  parser.ItemPlus,
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
  "IN":       parser.ItemIn,
  "MACRO":    parser.ItemMacro,
  "BLOCK":    parser.ItemBlock,
  "END":      parser.ItemEnd,
}

type Lexer struct {
  *parser.Lexer
}

type TTerse struct {
  lexer *Lexer
  items []parser.LexItem
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

func (p *TTerse) next() parser.LexItem {
  return p.lexer.NextItem()
}

func (p *TTerse) NextItem() parser.LexItem {
  if len(p.items) > 0 {
    item := p.items[0]
    p.items = p.items[1:]
    return item
  }
  return p.next()
}

func (p *TTerse) NextNonSpaceItem() parser.LexItem {
  for {
    n := p.NextItem()
    switch n.Type() {
    case parser.ItemEOF, parser.ItemError:
      return parser.NewLexItem(parser.ItemEOF, 0, "")
    case parser.ItemSpace:
      continue
    default:
      return n
    }
  }
}

func (p *TTerse) Peek() parser.LexItem {
  item := p.NextNonSpaceItem()
  p.items = append(p.items, item)
  return item
}

func (p *TTerse) Parse(input string) (*parser.AST, error) {
  b := parser.NewBuilder()
  lex := NewLexer()
  lex.SetInput(input)
  return b.Parse("foo", input, lex)
}
/*

  p.lexer.SetInput(input)
  go p.lexer.Run()

Loop:
  for {
    item := p.NextNonSpaceItem()
    fmt.Printf("item -> %v\n", item)
    switch item.Type() {
    case parser.ItemEnd, parser.ItemEOF:
      break Loop
    case parser.ItemIdentifier:
      next := p.Peek()
      if next.Type() == parser.ItemAssign {
        p.ParseAssignment(item)
      }
    case parser.ItemAssign:
      // st.stack[st.curstack]
    case parser.ItemRawString:
    case parser.ItemSpace, parser.ItemTagStart, parser.ItemTagEnd:
      // Nothing to do
    }
  }

  return nil, nil
}
*/
