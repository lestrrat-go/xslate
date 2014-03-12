package parser

import (
  "fmt"
)
type Builder struct {

}

type BuilderCtx struct{
  ParseName string
  Text      string
  Lexer     LexRunner
  Root      *ListNode
  PeekCount int
  Tokens    [3]LexItem
  ParentStack []*ListNode
  CurrentStackTop int
}

func NewBuilder() *Builder {
  return &Builder {}
}

func (b *Builder) Parse(name, text string, lex LexRunner) (*AST, error) {
  // defer b.recover
  ctx := &BuilderCtx {
    ParseName:  name,
    Text:       text,
    Lexer:      lex,
    Root:       NewRootNode(),
    Tokens:     [3]LexItem {},
    CurrentStackTop: -1,
    ParentStack:[]*ListNode {},
  }
  b.Start(ctx)
  b.ParseStatements(ctx)
  return &AST { Root: ctx.Root }, nil
}

func (b *Builder) Start(ctx *BuilderCtx) {
  go ctx.Lexer.Run()
}

func (b *Builder) Backup(ctx *BuilderCtx) {
  ctx.PeekCount++
}

func (b *Builder) Peek(ctx *BuilderCtx) LexItem {
  if ctx.PeekCount > 0 {
    return ctx.Tokens[ctx.PeekCount - 1]
  }
  ctx.PeekCount = 1
  ctx.Tokens[0] = ctx.Lexer.NextItem()
  return ctx.Tokens[0]
}

func (b *Builder) PeekNonSpace(ctx *BuilderCtx) LexItem {
  var token LexItem
  for {
    token = b.Next(ctx)
    if token.Type() != ItemSpace {
      break
    }
  }
  b.Backup(ctx)
  return token
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

func (b *Builder) Backup2(ctx *BuilderCtx, t1 LexItem) {
  ctx.Tokens[1] = t1
  ctx.PeekCount = 2
}

func (ctx *BuilderCtx) PushParentNode(n *ListNode) {
  ctx.CurrentStackTop++
  if ctx.CurrentStackTop >= len(ctx.ParentStack) {
    ctx.ParentStack = append(ctx.ParentStack, n)
  } else {
    ctx.ParentStack[ctx.CurrentStackTop] = n
  }
}

func (ctx *BuilderCtx) PopParentNode() Node {
  cur := ctx.CurrentParentNode()
  ctx.CurrentStackTop--
  return cur
}

func (ctx *BuilderCtx) CurrentParentNode() *ListNode {
  if ctx.CurrentStackTop >= 0 {
    return ctx.ParentStack[ctx.CurrentStackTop]
  }
  return nil
}

func (b *Builder) ParseStatements(ctx *BuilderCtx) Node {
  ctx.PushParentNode(ctx.Root)
  for b.Peek(ctx).Type() != ItemEOF {
    n := b.ParseTemplateOrText(ctx)
    if n != nil {
      ctx.CurrentParentNode().Append(n)
    }
  }
  return nil
}

func (b *Builder) ParseTemplateOrText(ctx *BuilderCtx) Node {
  switch token := b.PeekNonSpace(ctx); token.Type() {
  case ItemRawString:
    b.NextNonSpace(ctx)
    return NewTextNode(token.Pos(), token.Value())
  case ItemTagStart:
    node := b.ParseTemplate(ctx)
    if node == nil {
      return nil
    }

    switch node.Type() {
    case NodeForeach, NodeWrapper:
      ctx.CurrentParentNode().Append(node)
      ctx.PushParentNode(node.(*ListNode))
      node = nil
    }
    return node
  default:
    panic(fmt.Sprintf("fuck %s", token))
  }
  return nil
}

func (b *Builder) Unexpected(format string, args ...interface{}) {
  panic(
    fmt.Sprintf(
      "Unexpected token found: %s",
      fmt.Sprintf(format, args...),
    ),
  )
}

func (b *Builder) ParseTemplate(ctx *BuilderCtx) Node {
  // consume tagstart
  start := b.NextNonSpace(ctx)
  if start.Type() != ItemTagStart {
    b.Unexpected("Expected TagStart, got %s", start)
  }

  var tmpl Node
  switch b.PeekNonSpace(ctx).Type() {
  case ItemEnd:
    b.NextNonSpace(ctx)
    parent := ctx.PopParentNode()
    if parent.Type() == NodeRoot {
      b.Unexpected("Unexpected END")
    }
  case ItemWrapper:
    tmpl = b.ParseWrapper(ctx)
  case ItemForeach:
    tmpl = b.ParseForeach(ctx)
  case ItemTagEnd: // Silly, but possible
    b.NextNonSpace(ctx)
    tmpl = NewNoopNode()
  case ItemIdentifier:
    tmpl = b.ParseExpression(ctx, true)
  default:
    b.Unexpected("%s", b.PeekNonSpace(ctx))
  }
  // Consume tag end
  end := b.NextNonSpace(ctx)
  if end.Type() != ItemTagEnd {
    b.Unexpected("Expected TagEnd, got %s", end)
  }
  return tmpl
}

func (b *Builder) ParseWrapper(ctx *BuilderCtx) Node {
  wrapper := b.Next(ctx)
  if wrapper.Type() != ItemWrapper {
    panic("fuck")
  }

  tmpl := b.NextNonSpace(ctx)
  switch tmpl.Type() {
  case ItemDoubleQuotedString, ItemSingleQuotedString:
    // no op
  default:
    b.Unexpected("Expected identifier, got %s", tmpl)
  }

  // If we have parameters, we have WITH. otherwise we want TagEnd
  switch token := b.PeekNonSpace(ctx); token.Type() {
  case ItemTagEnd:
    return NewWrapperNode(token.Pos(), tmpl.Value())
  case ItemWith:
    b.NextNonSpace(ctx) // WITH
    wrapper := NewWrapperNode(token.Pos(), tmpl.Value())

    for {
      token := b.PeekNonSpace(ctx)
      if token.Type() == ItemTagEnd {
        break
      }

      assignment := b.ParseAssignment(ctx)
      wrapper.Append(assignment)

      if b.PeekNonSpace(ctx).Type() != ItemComma {
        break
      }

      // comma
      b.NextNonSpace(ctx)
    }
    return wrapper
  default:
    panic( b.PeekNonSpace(ctx).Type() )
  }
  return nil
}

func (b *Builder) ParseAssignment(ctx *BuilderCtx) Node {
  symbol := b.NextNonSpace(ctx)
  if symbol.Type() != ItemIdentifier {
    b.Unexpected("Expected identifier, got %s", symbol)
  }

  eq := b.NextNonSpace(ctx)
  if eq.Type() != ItemAssign {
    b.Unexpected("Expected assign, got %s", eq)
  }

  node := NewAssignmentNode(symbol.Pos(), symbol.Value())
  node.Append(b.ParseExpression(ctx, false))
  return node
}

func (b *Builder) ParseExpression(ctx *BuilderCtx, canPrint bool) Node {
  token := b.PeekNonSpace(ctx)
  switch token.Type() {
  case ItemIdentifier:
    // Could be a single var, or a method call
    b.NextNonSpace(ctx)
    next := b.PeekNonSpace(ctx)
    switch next.Type() {
    case ItemPeriod:
      // if an identifier is followed by a period, it's either
      // a method call or a variable fetch. 
      b.NextNonSpace(ctx)
      n := b.ParseMethodOrFetch(ctx, token)
      if canPrint {
        return NewPrintNode(next.Pos(), n)
      }
      return n
    case ItemAssign:
      b.Backup2(ctx, token)
      return b.ParseAssignment(ctx)
    case ItemTagEnd:
      n := NewLocalVarNode(token.Pos(), token.Value())
      if canPrint {
        return NewPrintNode(token.Pos(), n)
      }
      return n
    default:
      b.Unexpected("Unknown token %s", next)
    }
  default:
    n := b.ParseLiteral(ctx)
    if canPrint {
      return NewPrintNode(token.Pos(), n)
    }
    return n
  }
  return nil
}

func (b *Builder) ParseMethodOrFetch(ctx *BuilderCtx, symbol LexItem) Node {
  // must find another identifier node
  next := b.NextNonSpace(ctx)
  if next.Type() != ItemIdentifier {
    b.Unexpected("Expected identifier, got %s", next)
  }

  paren := b.PeekNonSpace(ctx)
  if paren.Type() != ItemOpenParen {
    return NewFetchFieldNode(symbol.Pos(), symbol.Value(), next.Value())
  }

  // Methodcall!
  b.NextNonSpace(ctx) // "("
  args := b.ParseExpression(ctx, false)
  paren = b.NextNonSpace(ctx)
  if paren.Type() != ItemCloseParen {
    b.Unexpected("Expected close parenthesis, got %s", paren)
  }
  return NewMethodcallNode(symbol.Pos(), symbol.Value(), next.Value(), args)
}

func (b *Builder) ParseLiteral(ctx *BuilderCtx) Node {
  t := b.NextNonSpace(ctx)
  switch t.Type() {
  case ItemDoubleQuotedString, ItemSingleQuotedString:
    return NewTextNode(t.Pos(), t.Value())
  case ItemNumber:
    return NewNumberNode(t.Pos(), t.Value())
  default:
    b.Unexpected("Expected literal value, got %s", t)
  }
  return nil
}

func (b *Builder) ParseForeach(ctx *BuilderCtx) Node {
  foreach := b.NextNonSpace(ctx)
  if foreach.Type() != ItemForeach {
    b.Unexpected("Expected FOREACH, got %s", foreach)
  }

  localsym := b.NextNonSpace(ctx)
  if localsym.Type() != ItemIdentifier {
    b.Unexpected("Expected identifier, got %s", localsym)
  }

  forNode := NewForeachNode(foreach.Pos(), localsym.Value())

  in := b.NextNonSpace(ctx)
  if in.Type() != ItemIn {
    b.Unexpected("Expected IN, got %s", in)
  }

  list := b.ParseList(ctx)
  forNode.Append(list)

  return forNode
}

func (b *Builder) ParseList(ctx *BuilderCtx) Node {
  list := b.NextNonSpace(ctx)
  if list.Type() != ItemIdentifier {
    b.Unexpected("Expected identifier, got %s", list)
  }

  return NewLocalVarNode(list.Pos(), list.Value())
}
