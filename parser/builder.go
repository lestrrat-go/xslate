package parser

import (
  "fmt"
  "strconv"
  "strings"
)
type Builder struct {

}

type StackFrame struct {
  Node        NodeAppender
  LocalVars   map[string]int
  LocalVarIdx int
}

type BuilderCtx struct{
  ParseName string
  Text      string
  Lexer     LexRunner
  Root      *ListNode
  PeekCount int
  Tokens    [3]LexItem
  ParentStack []*StackFrame
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
    ParentStack:[]*StackFrame {},
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

func (ctx *BuilderCtx) HasLocalVar(symbol string) (int, bool) {
  frame := ctx.CurrentStackFrame()
  pos, ok := frame.LocalVars[symbol]
  return pos, ok
}

func (ctx *BuilderCtx) DeclareLocalVar(symbol string) int {
  frame := ctx.CurrentStackFrame()

  frame.LocalVars[symbol] = frame.LocalVarIdx
  frame.LocalVarIdx++

  return frame.LocalVarIdx - 1
}

func (ctx *BuilderCtx) PushStackFrame() *StackFrame {
  frame := &StackFrame {
    LocalVars: make(map[string]int),
    LocalVarIdx: 0,
  }

  ctx.CurrentStackTop++
  if ctx.CurrentStackTop >= len(ctx.ParentStack) {
    ctx.ParentStack = append(ctx.ParentStack, frame)
  } else {
    ctx.ParentStack[ctx.CurrentStackTop] = frame
  }

  return frame
}

func (ctx *BuilderCtx) PopStackFrame() *StackFrame {
  n := ctx.ParentStack[ctx.CurrentStackTop]
  ctx.CurrentStackTop--
  return n
}

func (ctx *BuilderCtx) CurrentStackFrame() *StackFrame {
  if ctx.CurrentStackTop > -1 {
    return ctx.ParentStack[ctx.CurrentStackTop]
  }
  return nil
}

func (ctx *BuilderCtx) PushParentNode(n NodeAppender) {
  frame := ctx.PushStackFrame()
  frame.Node = n
}

func (ctx *BuilderCtx) PopParentNode() NodeAppender {
  frame := ctx.PopStackFrame()
  if frame != nil {
    return frame.Node
  }
  return nil
}

func (ctx *BuilderCtx) CurrentParentNode() NodeAppender {
  frame := ctx.CurrentStackFrame()
  if frame != nil {
    return frame.Node
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
    return b.ParseRawString(ctx)
  case ItemTagStart:
    return b.ParseTemplate(ctx)
  default:
    panic(fmt.Sprintf("fuck %s", token))
  }
  return nil
}

func (b *Builder) ParseRawString(ctx *BuilderCtx) Node {
  token := b.NextNonSpace(ctx)
  n := NewPrintRawNode(token.Pos())
  n.Append(NewTextNode(token.Pos(), token.Value()))
  return n
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
    for keepPopping := true; keepPopping; {
      parent := ctx.PopParentNode()
      switch parent.Type() {
      case NodeRoot:
        b.Unexpected("Unexpected END")
      case NodeElse:
        // no op
      default:
        keepPopping = false
      }
    }
  case ItemComment:
    b.NextNonSpace(ctx)
    // no op
  case ItemSet:
    b.NextNonSpace(ctx)
    tmpl = b.ParseAssignment(ctx)
  case ItemWrapper:
    tmpl = b.ParseWrapper(ctx)
  case ItemForeach:
    tmpl = b.ParseForeach(ctx)
  case ItemInclude:
    tmpl = b.ParseInclude(ctx)
  case ItemTagEnd: // Silly, but possible
    b.NextNonSpace(ctx)
    tmpl = NewNoopNode()
  case ItemIdentifier, ItemNumber, ItemDoubleQuotedString, ItemSingleQuotedString, ItemOpenParen:
    tmpl = b.ParseExpression(ctx, true)
  case ItemIf:
    tmpl = b.ParseIf(ctx)
  case ItemElse:
    tmpl = b.ParseElse(ctx)
  default:
    b.Unexpected("%s", b.PeekNonSpace(ctx))
  }

  for b.PeekNonSpace(ctx).Type() == ItemComment {
    b.NextNonSpace(ctx)
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
  var template string
  switch tmpl.Type() {
  case ItemDoubleQuotedString, ItemSingleQuotedString:
    template = tmpl.Value()
    template = template[1:len(template)-1]
  default:
    b.Unexpected("Expected identifier, got %s", tmpl)
  }

  n := NewWrapperNode(wrapper.Pos(), template)
  ctx.CurrentParentNode().Append(n)
  ctx.PushParentNode(n)

  ctx.PushStackFrame()

  // If we have parameters, we have WITH. otherwise we want TagEnd
  if token := b.PeekNonSpace(ctx); token.Type() != ItemWith {
    ctx.PopStackFrame()
    return nil
  }
  b.NextNonSpace(ctx) // WITH
  for {
    a := b.ParseAssignment(ctx)
    n.AppendAssignment(a)
    next := b.PeekNonSpace(ctx)
    if next.Type() != ItemComma {
      break
    } else if  next.Type() == ItemTagEnd {
      break
    }
    b.NextNonSpace(ctx)
  }
  ctx.PopStackFrame()

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
  node.Expression = b.ParseExpression(ctx, false)

  ctx.DeclareLocalVar(symbol.Value())
  return node
}

func (b *Builder) LocalVarOrFetchSymbol(ctx *BuilderCtx, token LexItem) Node {
  if idx, ok := ctx.HasLocalVar(token.Value()); ok {
    return NewLocalVarNode(token.Pos(), token.Value(), idx)
  }
  return NewFetchSymbolNode(token.Pos(), token.Value())
}

func (b *Builder) ParseTerm(ctx *BuilderCtx) Node {
  switch token := b.NextNonSpace(ctx); token.Type() {
  case ItemIdentifier:
    return b.LocalVarOrFetchSymbol(ctx, token)
  case ItemNumber, ItemDoubleQuotedString, ItemSingleQuotedString:
    b.Backup(ctx)
    return b.ParseLiteral(ctx)
  default:
    b.Backup(ctx)
    return nil
  }
}

func (b *Builder) ParseFunCall(ctx *BuilderCtx, invocant Node) Node {
  next := b.NextNonSpace(ctx)
  if next.Type() != ItemOpenParen {
    b.Unexpected("Expected '(', got %s", next.Type())
  }

  args := b.ParseList(ctx)
  closeParen := b.NextNonSpace(ctx)
  if closeParen.Type() != ItemCloseParen {
    b.Unexpected("Expected ')', got %s", closeParen.Type())
  }
  return NewFunCallNode(invocant.Position(), invocant, args.(*ListNode))
}

func (b *Builder) ParseMethodCallOrMapLookup(ctx *BuilderCtx, invocant Node) Node {
  // We have already seen identifier followed by a period
  symbol := b.NextNonSpace(ctx)
  if symbol.Type() != ItemIdentifier {
    b.Unexpected("Expected identifier for method call or map lookup, got %s", symbol.Type())
  }

  var n Node
  next := b.NextNonSpace(ctx)
  if next.Type() != ItemOpenParen {
    // it's a map lookup. Put back that extra token we read
    b.Backup(ctx)
    n = NewFetchFieldNode(invocant.Position(), invocant, symbol.Value())
  } else {
    // It's a method call! Parse the list
    args := b.ParseList(ctx)
    closeParen := b.NextNonSpace(ctx)
    if closeParen.Type() != ItemCloseParen {
      b.Unexpected("Expected ')', got %s", closeParen.Type())
    }
    n = NewMethodCallNode(invocant.Position(), invocant, symbol.Value(), args.(*ListNode))
  }

  // If we are followed by another period, we are going to have to
  // check for another level of methodcall / lookup
  if b.PeekNonSpace(ctx).Type() == ItemPeriod {
    b.NextNonSpace(ctx) // consume period
    return b.ParseMethodCallOrMapLookup(ctx, n)
  }
  return n
}

func (b *Builder) ParseExpression(ctx *BuilderCtx, canPrint bool) (n Node) {
  defer func() {
    if n != nil && canPrint {
      n = NewPrintNode(n.Position(), n)
    }
  }()

  if b.PeekNonSpace(ctx).Type() == ItemOpenParen {
    n = b.ParseGroup(ctx)
  } else {
    n = b.ParseTerm(ctx)
    if n == nil {
      panic("TODO")
    }
  }

  next := b.NextNonSpace(ctx);

  switch n.Type() {
  case NodeLocalVar, NodeFetchSymbol:
    switch next.Type() {
    case ItemPeriod:
      // It's either a method call, or a map lookup
      n = b.ParseMethodCallOrMapLookup(ctx, n)
    case ItemOpenParen:
      b.Backup(ctx) // put back the open paren
      // A variable followed by an open paren is a function call
      n = b.ParseFunCall(ctx, n)
    default:
      b.Backup(ctx)
    }
  default:
    b.Backup(ctx)
  }

  next = b.NextNonSpace(ctx)
  switch next.Type() {
  case ItemPlus:
    tmp := NewPlusNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
    return
  case ItemMinus:
    tmp := NewMinusNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
    return
  case ItemAsterisk:
    tmp := NewMulNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
    return
  case ItemSlash:
    tmp := NewDivNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
    return
  case ItemLT:
    tmp := NewLTNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
    return
  case ItemGT:
    tmp := NewGTNode(next.Pos())
    tmp.Left = n
    tmp.Right = b.ParseExpression(ctx, false)
    n = tmp
  case ItemVerticalSlash:
    b.Backup(ctx)
    n = b.ParseFilter(ctx, n)
  default:
    b.Backup(ctx)
  }

  return
}

func (b *Builder) ParseFilter(ctx *BuilderCtx, n Node) Node {
  vslash := b.NextNonSpace(ctx)
  if vslash.Type() != ItemVerticalSlash {
    b.Unexpected("Expected '|', got %s", vslash.Type())
  }

  id := b.NextNonSpace(ctx)
  if id.Type() != ItemIdentifier {
    b.Unexpected("Expected idenfitier, got %s", id.Type())
  }

  filter := NewFilterNode(id.Pos(), id.Value(), n)

  if b.PeekNonSpace(ctx).Type() == ItemVerticalSlash {
    filter = b.ParseFilter(ctx, filter).(*FilterNode)
  }

  return filter
}

func (b *Builder) ParseLiteral(ctx *BuilderCtx) Node {
  t := b.NextNonSpace(ctx)
  switch t.Type() {
  case ItemDoubleQuotedString, ItemSingleQuotedString:
    v := t.Value()
    return NewTextNode(t.Pos(), v[1:len(v) - 1])
  case ItemNumber:
    v := t.Value()
    // XXX TODO: parse hex/oct/bin
    if strings.Contains(v, ".") {
      f, err := strconv.ParseFloat(v, 64)
      if err != nil { // shouldn't happen, as we were able to lex it
        b.Unexpected("Could not parse number: %s", err)
      }
      return NewFloatNode(t.Pos(), f)
    }
    i, err := strconv.ParseInt(v, 10, 64)
    if err != nil {
      b.Unexpected("Could not parse number: %s", err)
    }
    return NewIntNode(t.Pos(), i)
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

  forNode.List = b.ParseListVariableOrMakeArray(ctx)

  ctx.CurrentParentNode().Append(forNode)
  ctx.PushParentNode(forNode)
  ctx.DeclareLocalVar(localsym.Value())

  return nil
}

func (b *Builder) ParseRange(ctx *BuilderCtx) Node {
  start := b.NextNonSpace(ctx)
  if start.Type() != ItemNumber {
    b.Unexpected("Expected number, got %s", start.Value())
  }

  rangeOp := b.NextNonSpace(ctx)
  if rangeOp.Type() != ItemRange {
    b.Unexpected("Expected range, got %s", rangeOp.Value())
  }

  end := b.NextNonSpace(ctx)
  if end.Type() != ItemNumber {
    b.Unexpected("Expected number, got %s", end.Value())
  }

  startN, _ := strconv.ParseInt(start.Value(), 10, 64)
  endN, _   := strconv.ParseInt(end.Value(), 10, 64)
  return NewRangeNode(start.Pos(), int(startN), int(endN))
}

func (b *Builder) ParseListVariableOrMakeArray(ctx *BuilderCtx) Node {
  list := b.PeekNonSpace(ctx)

  var n Node
  switch list.Type() {
  case ItemIdentifier:
    b.NextNonSpace(ctx)
    if idx, ok := ctx.HasLocalVar(list.Value()); ok {
      n = NewLocalVarNode(list.Pos(), list.Value(), idx)
    } else {
      n = NewFetchSymbolNode(list.Pos(), list.Value())
    }
    if b.PeekNonSpace(ctx).Type() == ItemPeriod {
      b.NextNonSpace(ctx)
      n = b.ParseMethodCallOrMapLookup(ctx, n)
    }
  case ItemOpenSquareBracket:
    n = b.ParseMakeArray(ctx)
  default:
    panic("fuck")
  }
  return n
}

func (b *Builder) ParseMakeArray(ctx *BuilderCtx) Node {
  openB := b.NextNonSpace(ctx)
  if openB.Type() != ItemOpenSquareBracket {
    b.Unexpected("Expected '[', got %s", openB.Value())
  }

  child := b.ParseList(ctx)

  closeB := b.NextNonSpace(ctx)
  if closeB.Type() != ItemCloseSquareBracket {
    b.Unexpected("Expected ']', got %s", closeB.Value())
  }

  return NewMakeArrayNode(openB.Pos(), child)
}

func (b *Builder) ParseList(ctx *BuilderCtx) Node {
  n := NewListNode(b.PeekNonSpace(ctx).Pos())
  OUTER: for {
    // At the beginning of this loop, we must see an
    // identifier or a literal
    switch item := b.PeekNonSpace(ctx); item.Type() {
    case ItemIdentifier, ItemNumber, ItemDoubleQuotedString, ItemSingleQuotedString:
      // okay, proceed
    default:
      break OUTER
    }

    // Depending on the next item, we have range operator or a literal list
    var child Node
    item := b.NextNonSpace(ctx)
    switch nextN := b.PeekNonSpace(ctx); nextN.Type() {
    case ItemRange:
      b.Backup2(ctx, item)
      child = b.ParseRange(ctx)
    default:
      b.Backup2(ctx, item)
      child = b.ParseExpression(ctx, false)
    }

    n.Append(child)

    // Then, we must be followed by either a comma, or the it's the end of the
    // list section
    if b.PeekNonSpace(ctx).Type() == ItemComma {
      b.NextNonSpace(ctx)
    }
  }
  return n
}

func (b *Builder) ParseIf(ctx *BuilderCtx) Node {
  ifToken := b.NextNonSpace(ctx)
  if ifToken.Type() != ItemIf {
    b.Unexpected("Expected if, got %s", ifToken)
  }

  // parenthesis are optional
  expectCloseParen := false
  if b.PeekNonSpace(ctx).Type() == ItemOpenParen {
    b.NextNonSpace(ctx)
    expectCloseParen = true
  }

  exp := b.ParseExpression(ctx, false)
  ifNode := NewIfNode(ifToken.Pos(), exp)

  if expectCloseParen {
    closeParenToken := b.NextNonSpace(ctx)
    if closeParenToken.Type() != ItemCloseParen {
      b.Unexpected("Expected close parenthesis, got %s", closeParenToken)
    }
  }

  ctx.CurrentParentNode().Append(ifNode)
  ctx.PushParentNode(ifNode)

  return nil
}

func (b *Builder) ParseElse(ctx *BuilderCtx) Node {
  elseToken := b.NextNonSpace(ctx)
  if elseToken.Type() != ItemElse {
    b.Unexpected("Expected else, got %s", elseToken)
  }

  // CurrentParentNode must be "If" in order for "else" to work
  if ctx.CurrentParentNode().Type() != NodeIf {
    b.Unexpected("Found else without if")
  }

  elseNode := NewElseNode(elseToken.Pos())
  elseNode.IfNode = ctx.CurrentParentNode()
  ctx.CurrentParentNode().Append(elseNode)
  ctx.PushParentNode(elseNode)

  return nil
}

func (b *Builder) ParseInclude(ctx *BuilderCtx) Node {
  incToken := b.NextNonSpace(ctx)
  if incToken.Type() != ItemInclude {
    b.Unexpected("Expected include, got %s", incToken)
  }

  // Next thing must be the name of the included template
  n := b.ParseExpression(ctx, false)
  x := NewIncludeNode(incToken.Pos(), n)
  ctx.PushStackFrame()

  if b.PeekNonSpace(ctx).Type() != ItemWith {
    ctx.PopStackFrame()
    return x
  }

  b.NextNonSpace(ctx)
  for {
    a := b.ParseAssignment(ctx)
    x.AppendAssignment(a)
    next := b.PeekNonSpace(ctx)
    if next.Type() != ItemComma {
      break
    } else if  next.Type() == ItemTagEnd {
      break
    }
    b.NextNonSpace(ctx)
  }
  ctx.PopStackFrame()

  return x
}

func (b *Builder) ParseGroup(ctx *BuilderCtx) Node {
  openParenToken := b.NextNonSpace(ctx)
  if openParenToken.Type() != ItemOpenParen {
    b.Unexpected("Expected '(', got %s", openParenToken)
  }

  n := NewGroupNode(openParenToken.Pos())
  n.Child = b.ParseExpression(ctx, false)

  closeParenToken := b.NextNonSpace(ctx)
  if closeParenToken.Type() != ItemCloseParen {
    b.Unexpected("Expected ')', got %s", closeParenToken)
  }

  return n
}
