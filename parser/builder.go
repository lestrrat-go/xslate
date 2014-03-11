package parser

type Builder struct {

}

type BuilderCtx struct{
  ParseName string
  Text      string
  Lexer     *Lexer
  Root      *ListNode
  PeekCount int
  Tokens    []LexItem
}

func NewBuilder() *Builder {
  return &Builder {}
}

func (b *Builder) Parse(name, text string) (*AST, error) {
  // defer b.recover
  lex := NewLexer()
  lex.SetInput(text)
  ctx := &BuilderCtx {
    ParseName:  name,
    Text:       text,
    Lexer:      lex,
    Root:       NewListNode(0),
  }
  b.Start(ctx)
  b.ParseStatements(ctx)
  return nil, nil
}

func (b *Builder) Start(ctx *BuilderCtx) {
  go ctx.Lexer.Run()
}

func (b *Builder) Peek(ctx *BuilderCtx) LexItem {
  if ctx.PeekCount > 0 {
    return ctx.Tokens[ctx.PeekCount - 1]
  }
  ctx.PeekCount =1
  ctx.Tokens[0] = ctx.Lexer.NextItem()
  return ctx.Tokens[0]
}

func (b *Builder) Next(ctx *BuilderCtx) LexItem {
  if ctx.PeekCount > 0 {
    ctx.PeekCount--
  } else {
    ctx.Tokens[0] = ctx.Lexer.NextItem()
  }
  return ctx.Tokens[ctx.PeekCount]
}

func (b *Builder) NextNonSpace(ctx *BuilderCtx) LexItem {
  var token LexItem
  for {
    token = b.Next(ctx)
    if token.Type() != ItemSpace {
      break
    }
  }
  return token
}

func (b *Builder) ParseStatements(ctx *BuilderCtx) Node {
  for b.Peek(ctx).Type() != ItemEOF {
    n := b.ParseTemplateOrText(ctx)
    if n != nil {
      ctx.Root.Append(n)
    }
  }
  return nil
}

func (b *Builder) ParseTemplateOrText(ctx *BuilderCtx) Node {
  switch token := b.NextNonSpace(ctx); token.Type() {
  case ItemRawString:
    return NewNode(NodeText, token.Pos(), token.Value())
  case ItemTagStart:
    return b.ParseTemplate(ctx)
  default:
    panic("fuck")
  }
  return nil
}

func (b *Builder) ParseTemplate(ctx *BuilderCtx) Node {
  return nil
}
