package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lestrrat/go-lex"
	"github.com/lestrrat/go-xslate/node"
	"github.com/lestrrat/go-xslate/util"
)

func NewFrame(s *util.Stack) *Frame {
	f := &Frame{
		util.NewFrame(s),
		nil,
		make(map[string]int),
	}
	return f
}

type builderCtx struct {
	ParseName       string
	Text            string
	Line            int
	Lexer           lex.Lexer
	Root            *node.ListNode
	PeekCount       int
	Tokens          [3]lex.LexItem
	CurrentStackTop int
	PostChomp       bool
	FrameStack      *util.Stack
	Frames          *util.Stack
	Error           error
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Parse(name string, l lex.Lexer) (ast *AST, err error) {
	ctx := &builderCtx{
		ParseName:  name,
		Lexer:      l,
		Root:       node.NewRootNode(),
		Tokens:     [3]lex.LexItem{},
		FrameStack: util.NewStack(5),
		Frames:     util.NewStack(5),
	}

	defer func() {
		if ctx.Error != nil {
			err = ctx.Error
			ast = nil
			// don't let the panic propagate
			recover()
		}
	}()

	b.Start(ctx)
	b.ParseStatements(ctx)
	return &AST{
		Name: name,
		Root: ctx.Root,
	}, nil
}

func (b *Builder) Start(ctx *builderCtx) {
	go ctx.Lexer.Run()
}

func (b *Builder) Backup(ctx *builderCtx) {
	ctx.PeekCount++
}

func (b *Builder) Peek(ctx *builderCtx) lex.LexItem {
	if ctx.PeekCount > 0 {
		return ctx.Tokens[ctx.PeekCount-1]
	}
	ctx.PeekCount = 1
	ctx.Tokens[0] = ctx.Lexer.NextItem()
	return ctx.Tokens[0]
}

func (b *Builder) PeekNonSpace(ctx *builderCtx) lex.LexItem {
	var token lex.LexItem
	for {
		token = b.Next(ctx)
		if token.Type() != ItemSpace {
			break
		}
	}
	b.Backup(ctx)
	return token
}

func (b *Builder) Next(ctx *builderCtx) lex.LexItem {
	if ctx.PeekCount > 0 {
		ctx.PeekCount--
	} else {
		ctx.Tokens[0] = ctx.Lexer.NextItem()
	}
	return ctx.Tokens[ctx.PeekCount]
}

func (b *Builder) NextNonSpace(ctx *builderCtx) lex.LexItem {
	var token lex.LexItem
	for {
		token = b.Next(ctx)
		ctx.Line = token.Line()
		if token.Type() != ItemSpace {
			break
		}
	}
	return token
}

func (b *Builder) Backup2(ctx *builderCtx, t1 lex.LexItem) {
	ctx.Tokens[1] = t1
	ctx.PeekCount = 2
}

func (ctx *builderCtx) HasLocalVar(symbol string) (pos int, ok bool) {
	for i := ctx.Frames.Cur(); i >= 0; i-- {
		frame, _ := ctx.Frames.Get(i)
		pos, ok = frame.(*Frame).LvarNames[symbol]
		if ok {
			return
		}
	}
	return 0, false
}

func (ctx *builderCtx) DeclareLocalVar(symbol string) int {
	frame := ctx.CurrentFrame()
	i := frame.DeclareVar(symbol)
	frame.LvarNames[symbol] = i
	return i
}

func (ctx *builderCtx) PushFrame() *Frame {
	f := NewFrame(ctx.FrameStack)
	ctx.Frames.Push(f)
	f.SetMark(ctx.Frames.Cur())
	return f
}

func (ctx *builderCtx) PopFrame() *Frame {
	x := ctx.Frames.Pop()
	if x == nil {
		return nil
	}

	f := x.(*Frame)
	for i := ctx.FrameStack.Cur(); i > f.Mark(); i-- {
		ctx.FrameStack.Pop()
	}
	return f
}

func (ctx *builderCtx) CurrentFrame() *Frame {
	x, err := ctx.Frames.Top()
	if err != nil {
		return nil
	}
	return x.(*Frame)
}

func (ctx *builderCtx) PushParentNode(n node.Appender) {
	frame := ctx.PushFrame()
	frame.Node = n
}

func (ctx *builderCtx) PopParentNode() node.Appender {
	frame := ctx.PopFrame()
	if frame != nil {
		return frame.Node
	}
	return nil
}

func (ctx *builderCtx) CurrentParentNode() node.Appender {
	frame := ctx.CurrentFrame()
	if frame != nil {
		return frame.Node
	}
	return nil
}

func (b *Builder) ParseStatements(ctx *builderCtx) node.Node {
	ctx.PushParentNode(ctx.Root)
	for b.Peek(ctx).Type() != ItemEOF {
		n := b.ParseTemplateOrText(ctx)
		if n != nil {
			ctx.CurrentParentNode().Append(n)
		}
	}
	return nil
}

func (b *Builder) ParseTemplateOrText(ctx *builderCtx) node.Node{
	token := b.PeekNonSpace(ctx)
	switch token.Type() {
	case ItemRawString:
		return b.ParseRawString(ctx)
	case ItemTagStart:
		return b.ParseTemplate(ctx)
	default:
		panic(fmt.Sprintf("Unexpected token: %s", token))
	}
}

func (b *Builder) ParseRawString(ctx *builderCtx) node.Node{
	const whiteSpace = " \t\r\n"
	token := b.NextNonSpace(ctx)
	if token.Type() != ItemRawString {
		b.Unexpected(ctx, "Expected raw string, got %s", token)
	}

	value := token.Value()

	if ctx.PostChomp {
		value = strings.TrimLeft(value, whiteSpace)
		ctx.PostChomp = false
	}

	// Look for signs of pre-chomp
	if b.PeekNonSpace(ctx).Type() == ItemTagStart {
		start := b.NextNonSpace(ctx)
		next := b.PeekNonSpace(ctx)
		b.Backup2(ctx, start)
		if next.Type() == ItemMinus {
			// prechomp!
			value = strings.TrimRight(value, whiteSpace)
		}
	}

	n := node.NewPrintRawNode(token.Pos())
	n.Append(node.NewTextNode(token.Pos(), value))

	return n
}

func (b *Builder) Unexpected(ctx *builderCtx, format string, args ...interface{}) {
	msg := fmt.Sprintf(
		"Unexpected token found: %s in %s at line %d",
		fmt.Sprintf(format, args...),
		ctx.ParseName,
		ctx.Line,
	)
	ctx.Error = errors.New(msg)
	panic(msg)
}

func (b *Builder) ParseTemplate(ctx *builderCtx) node.Node{
	// consume tagstart
	start := b.NextNonSpace(ctx)
	if start.Type() != ItemTagStart {
		b.Unexpected(ctx, "Expected TagStart, got %s", start)
	}
	ctx.PostChomp = false

	if b.PeekNonSpace(ctx).Type() == ItemMinus {
		b.NextNonSpace(ctx)
	}

	var tmpl node.Node
	switch b.PeekNonSpace(ctx).Type() {
	case ItemEnd:
		b.NextNonSpace(ctx)
		for keepPopping := true; keepPopping; {
			parent := ctx.PopParentNode()
			switch parent.Type() {
			case node.Root:
				b.Unexpected(ctx, "Unexpected END")
			case node.Else:
				// no op
			default:
				keepPopping = false
			}
		}
	case ItemComment:
		b.NextNonSpace(ctx)
		// no op
	case ItemCall:
		b.NextNonSpace(ctx)
		tmpl = b.ParseExpressionOrAssignment(ctx, false)
	case ItemSet:
		b.NextNonSpace(ctx) // Consume SET
		tmpl = b.ParseAssignment(ctx)
	case ItemMacro:
		tmpl = b.ParseMacro(ctx)
	case ItemWrapper:
		tmpl = b.ParseWrapper(ctx)
	case ItemForeach:
		tmpl = b.ParseForeach(ctx)
	case ItemWhile:
		tmpl = b.ParseWhile(ctx)
	case ItemInclude:
		tmpl = b.ParseInclude(ctx)
	case ItemTagEnd: // Silly, but possible
		b.NextNonSpace(ctx)
		tmpl = node.NewNoopNode()
	case ItemIdentifier, ItemNumber, ItemDoubleQuotedString, ItemSingleQuotedString, ItemOpenParen:
		tmpl = b.ParseExpressionOrAssignment(ctx, true)
	case ItemIf:
		tmpl = b.ParseIf(ctx)
	case ItemElse:
		tmpl = b.ParseElse(ctx)
	default:
		b.Unexpected(ctx, "%s", b.PeekNonSpace(ctx))
	}

	for b.PeekNonSpace(ctx).Type() == ItemComment {
		b.NextNonSpace(ctx)
	}

	if b.PeekNonSpace(ctx).Type() == ItemMinus {
		b.NextNonSpace(ctx)
		ctx.PostChomp = true
	}

	// Consume tag end
	end := b.NextNonSpace(ctx)
	if end.Type() != ItemTagEnd {
		b.Unexpected(ctx, "Expected TagEnd, got %s", end)
	}
	return tmpl
}

func (b *Builder) ParseExpressionOrAssignment(ctx *builderCtx, canPrint bool) node.Node{
	// There's a special case for assignment where SET is omitted
	// [% foo = ... %] instead of [% SET foo = ... %]
	next := b.NextNonSpace(ctx)
	following := b.PeekNonSpace(ctx)
	b.Backup2(ctx, next)

	var n node.Node
	if next.Type() == ItemIdentifier {
		switch following.Type() {
		case ItemAssign, ItemAssignAdd, ItemAssignSub, ItemAssignMul, ItemAssignDiv:
			// This is a simple assignment!
			n = b.ParseAssignment(ctx)
		default:
			n = b.ParseExpression(ctx, canPrint)
		}
	} else {
		n = b.ParseExpression(ctx, canPrint)
	}

	return n
}

func (b *Builder) ParseWrapper(ctx *builderCtx) node.Node{
	wrapper := b.Next(ctx)
	if wrapper.Type() != ItemWrapper {
		panic("fuck")
	}

	tmpl := b.NextNonSpace(ctx)
	var template string
	switch tmpl.Type() {
	case ItemDoubleQuotedString, ItemSingleQuotedString:
		template = tmpl.Value()
		template = template[1 : len(template)-1]
	default:
		b.Unexpected(ctx, "Expected identifier, got %s", tmpl)
	}

	n := node.NewWrapperNode(wrapper.Pos(), template)
	ctx.CurrentParentNode().Append(n)
	ctx.PushParentNode(n)

	ctx.PushFrame()

	// If we have parameters, we have WITH. otherwise we want TagEnd
	if token := b.PeekNonSpace(ctx); token.Type() != ItemWith {
		ctx.PopFrame()
		return nil
	}
	b.NextNonSpace(ctx) // WITH
LOOP:
	for {
		a := b.ParseAssignment(ctx)
		n.AppendAssignment(a)
		next := b.PeekNonSpace(ctx)
		switch next.Type() {
		case ItemComma, ItemTagEnd:
			break LOOP
		case ItemMinus:
			cur := b.NextNonSpace(ctx)
			next := b.PeekNonSpace(ctx)
			b.Backup2(ctx, cur)
			if next.Type() == ItemTagEnd {
				break LOOP
			}
		}
		b.NextNonSpace(ctx)
	}
	ctx.PopFrame()

	return nil
}

func (b *Builder) ParseAssignment(ctx *builderCtx) node.Node{
	symbol := b.NextNonSpace(ctx)
	if symbol.Type() != ItemIdentifier {
		b.Unexpected(ctx, "Expected identifier, got %s", symbol)
	}

	b.DeclareLocalVarIfNew(ctx, symbol)
	n := node.NewAssignmentNode(symbol.Pos(), symbol.Value())

	eq := b.NextNonSpace(ctx)
	switch eq.Type() {
	case ItemAssign:
		n.Expression = b.ParseExpression(ctx, false)
	case ItemAssignAdd:
		add := node.NewPlusNode(symbol.Pos())
		add.Left = b.LocalVarOrFetchSymbol(ctx, symbol)
		add.Right = b.ParseExpression(ctx, false)
		n.Expression = add
	case ItemAssignSub:
		sub := node.NewMinusNode(symbol.Pos())
		sub.Left = b.LocalVarOrFetchSymbol(ctx, symbol)
		sub.Right = b.ParseExpression(ctx, false)
		n.Expression = sub
	case ItemAssignMul:
		mul := node.NewMulNode(symbol.Pos())
		mul.Left = b.LocalVarOrFetchSymbol(ctx, symbol)
		mul.Right = b.ParseExpression(ctx, false)
		n.Expression = mul
	case ItemAssignDiv:
		div := node.NewDivNode(symbol.Pos())
		div.Left = b.LocalVarOrFetchSymbol(ctx, symbol)
		div.Right = b.ParseExpression(ctx, false)
		n.Expression = div
	default:
		b.Unexpected(ctx, "Expected assign, got %s", eq)
	}
	return n
}

func (b *Builder) DeclareLocalVarIfNew(ctx *builderCtx, symbol lex.LexItem) {
	if _, ok := ctx.HasLocalVar(symbol.Value()); !ok {
		ctx.DeclareLocalVar(symbol.Value())
	}
}

func (b *Builder) LocalVarOrFetchSymbol(ctx *builderCtx, token lex.LexItem) node.Node{
	if idx, ok := ctx.HasLocalVar(token.Value()); ok {
		return node.NewLocalVarNode(token.Pos(), token.Value(), idx)
	}
	return node.NewFetchSymbolNode(token.Pos(), token.Value())
}

func (b *Builder) ParseTerm(ctx *builderCtx) node.Node{
	token := b.NextNonSpace(ctx)
	switch token.Type() {
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

func (b *Builder) ParseFunCall(ctx *builderCtx, invocant node.Node) node.Node {
	next := b.NextNonSpace(ctx)
	if next.Type() != ItemOpenParen {
		b.Unexpected(ctx, "Expected '(', got %s", next.Type())
	}

	args := b.ParseList(ctx)
	closeParen := b.NextNonSpace(ctx)
	if closeParen.Type() != ItemCloseParen {
		b.Unexpected(ctx, "Expected ')', got %s", closeParen.Type())
	}

	return node.NewFunCallNode(invocant.Pos(), invocant, args.(*node.ListNode))
}

func (b *Builder) ParseMethodCallOrMapLookup(ctx *builderCtx, invocant node.Node) node.Node{
	// We have already seen identifier followed by a period
	symbol := b.NextNonSpace(ctx)
	if symbol.Type() != ItemIdentifier {
		b.Unexpected(ctx, "Expected identifier for method call or map lookup, got %s", symbol.Type())
	}

	var n node.Node
	next := b.NextNonSpace(ctx)
	if next.Type() != ItemOpenParen {
		// it's a map lookup. Put back that extra token we read
		b.Backup(ctx)
		n = node.NewFetchFieldNode(invocant.Pos(), invocant, symbol.Value())
	} else {
		// It's a method call! Parse the list
		args := b.ParseList(ctx)
		closeParen := b.NextNonSpace(ctx)
		if closeParen.Type() != ItemCloseParen {
			b.Unexpected(ctx, "Expected ')', got %s", closeParen.Type())
		}
		n = node.NewMethodCallNode(invocant.Pos(), invocant, symbol.Value(), args.(*node.ListNode))
	}

	// If we are followed by another period, we are going to have to
	// check for another level of methodcall / lookup
	if b.PeekNonSpace(ctx).Type() == ItemPeriod {
		b.NextNonSpace(ctx) // consume period
		return b.ParseMethodCallOrMapLookup(ctx, n)
	}
	return n
}

func (b *Builder) ParseArrayElementFetch(ctx *builderCtx, invocant node.Node) node.Node{
	openBracket := b.NextNonSpace(ctx)
	if openBracket.Type() != ItemOpenSquareBracket {
		b.Unexpected(ctx, "Expected '[', got %s", openBracket)
	}

	index := b.ParseExpression(ctx, false)

	n := node.NewFetchArrayElementNode(openBracket.Pos())
	n.Left = invocant
	n.Right = index

	closeBracket := b.NextNonSpace(ctx)
	if closeBracket.Type() != ItemCloseSquareBracket {
		b.Unexpected(ctx, "Expected ']', got %s", closeBracket)
	}

	return n
}

func (b *Builder) ParseExpression(ctx *builderCtx, canPrint bool) (n node.Node) {
	defer func() {
		if n != nil && canPrint {
			n = node.NewPrintNode(n.Pos(), n)
		}
	}()

	switch b.PeekNonSpace(ctx).Type() {
	case ItemOpenParen:
		// Looks like a group of something
		n = b.ParseGroup(ctx)
	case ItemOpenSquareBracket:
		// Looks like an inline list def
		n = b.ParseMakeArray(ctx)
	default:
		// Otherwise it's a straight forward ... something
		n = b.ParseTerm(ctx)
		if n == nil {
			panic(fmt.Sprintf("Expected term but could not parse. Next is %s\n", b.PeekNonSpace(ctx)))
		}
	}

	next := b.PeekNonSpace(ctx)

	switch n.Type() {
	case node.LocalVar, node.FetchSymbol:
		switch next.Type() {
		case ItemPeriod:
			// It's either a method call, or a map lookup
			b.NextNonSpace(ctx)
			n = b.ParseMethodCallOrMapLookup(ctx, n)
		case ItemOpenSquareBracket:
			n = b.ParseArrayElementFetch(ctx, n)
		case ItemOpenParen:
			// A variable followed by an open paren is a function call
			n = b.ParseFunCall(ctx, n)
		}
	}

	next = b.NextNonSpace(ctx)
	switch next.Type() {
	case ItemPlus:
		tmp := node.NewPlusNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
		return
	case ItemMinus:
		// This is special...
		following := b.PeekNonSpace(ctx)
		if following.Type() == ItemTagEnd {
			b.Backup2(ctx, next)
			// Postchomp! not arithmetic!
			return
		}
		tmp := node.NewMinusNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
		return
	case ItemAsterisk:
		tmp := node.NewMulNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
		return
	case ItemSlash:
		tmp := node.NewDivNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
		return
	case ItemEquals:
		tmp := node.NewEqualsNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
	case ItemNotEquals:
		tmp := node.NewNotEqualsNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
	case ItemLT:
		tmp := node.NewLTNode(next.Pos())
		tmp.Left = n
		tmp.Right = b.ParseExpression(ctx, false)
		n = tmp
		return
	case ItemGT:
		tmp := node.NewGTNode(next.Pos())
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

func (b *Builder) ParseFilter(ctx *builderCtx, n node.Node) node.Node{
	vslash := b.NextNonSpace(ctx)
	if vslash.Type() != ItemVerticalSlash {
		b.Unexpected(ctx, "Expected '|', got %s", vslash.Type())
	}

	id := b.NextNonSpace(ctx)
	if id.Type() != ItemIdentifier {
		b.Unexpected(ctx, "Expected idenfitier, got %s", id.Type())
	}

	filter := node.NewFilterNode(id.Pos(), id.Value(), n)

	if b.PeekNonSpace(ctx).Type() == ItemVerticalSlash {
		filter = b.ParseFilter(ctx, filter).(*node.FilterNode)
	}

	return filter
}

func (b *Builder) ParseLiteral(ctx *builderCtx) node.Node{
	t := b.NextNonSpace(ctx)
	switch t.Type() {
	case ItemDoubleQuotedString, ItemSingleQuotedString:
		v := t.Value()
		return node.NewTextNode(t.Pos(), v[1:len(v)-1])
	case ItemNumber:
		v := t.Value()
		// XXX TODO: parse hex/oct/bin
		if strings.Contains(v, ".") {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil { // shouldn't happen, as we were able to lex it
				b.Unexpected(ctx, "Could not parse number: %s", err)
			}
			return node.NewFloatNode(t.Pos(), f)
		}
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			b.Unexpected(ctx, "Could not parse number: %s", err)
		}
		return node.NewIntNode(t.Pos(), i)
	default:
		b.Unexpected(ctx, "Expected literal value, got %s", t)
	}
	return nil
}

func (b *Builder) ParseForeach(ctx *builderCtx) node.Node{
	foreach := b.NextNonSpace(ctx)
	if foreach.Type() != ItemForeach {
		b.Unexpected(ctx, "Expected FOREACH, got %s", foreach)
	}

	localsym := b.NextNonSpace(ctx)
	if localsym.Type() != ItemIdentifier {
		b.Unexpected(ctx, "Expected identifier, got %s", localsym)
	}

	forNode := node.NewForeachNode(foreach.Pos(), localsym.Value())
	// use current frame mark
	forNode.IndexVarIdx = ctx.CurrentFrame().Mark()

	in := b.NextNonSpace(ctx)
	if in.Type() != ItemIn {
		b.Unexpected(ctx, "Expected IN, got %s", in)
	}

	forNode.List = b.ParseListVariableOrMakeArray(ctx)

	ctx.CurrentParentNode().Append(forNode)
	ctx.PushParentNode(forNode)
	ctx.DeclareLocalVar(localsym.Value())
	ctx.DeclareLocalVar("loop")

	return nil
}

func (b *Builder) ParseWhile(ctx *builderCtx) node.Node{
	while := b.NextNonSpace(ctx)
	if while.Type() != ItemWhile {
		b.Unexpected(ctx, "Expected WHILE, got %s", while)
	}

	condition := b.ParseExpression(ctx, false)

	whileNode := node.NewWhileNode(while.Pos(), condition)

	ctx.CurrentParentNode().Append(whileNode)
	ctx.PushParentNode(whileNode)
	ctx.DeclareLocalVar("loop")

	return nil
}

func (b *Builder) ParseRange(ctx *builderCtx) node.Node{
	start := b.ParseTerm(ctx)
	if start == nil {
		b.Unexpected(ctx, "Expected term (start range), got %s", b.PeekNonSpace(ctx).Value())
	}

	rangeOp := b.NextNonSpace(ctx)
	if rangeOp.Type() != ItemRange {
		b.Unexpected(ctx, "Expected range, got %s", rangeOp.Value())
	}

	end := b.ParseTerm(ctx)
	if end == nil {
		b.Unexpected(ctx, "Expected term (end range), got %s", b.PeekNonSpace(ctx).Value())
	}

	return node.NewRangeNode(start.Pos(), start, end)
}

func (b *Builder) ParseListVariableOrMakeArray(ctx *builderCtx) node.Node{
	list := b.PeekNonSpace(ctx)

	var n node.Node
	switch list.Type() {
	case ItemIdentifier:
		b.NextNonSpace(ctx)
		if idx, ok := ctx.HasLocalVar(list.Value()); ok {
			n = node.NewLocalVarNode(list.Pos(), list.Value(), idx)
		} else {
			n = node.NewFetchSymbolNode(list.Pos(), list.Value())
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

func (b *Builder) ParseMakeArray(ctx *builderCtx) node.Node{
	openB := b.NextNonSpace(ctx)
	if openB.Type() != ItemOpenSquareBracket {
		b.Unexpected(ctx, "Expected '[', got %s", openB.Value())
	}

	child := b.ParseList(ctx)

	closeB := b.NextNonSpace(ctx)
	if closeB.Type() != ItemCloseSquareBracket {
		b.Unexpected(ctx, "Expected ']', got %s", closeB.Value())
	}

	return node.NewMakeArrayNode(openB.Pos(), child)
}

func (b *Builder) ParseList(ctx *builderCtx) node.Node{
	n := node.NewListNode(b.PeekNonSpace(ctx).Pos())
OUTER:
	for {
		// At the beginning of this loop, we must see an
		// identifier or a literal
		switch item := b.PeekNonSpace(ctx); item.Type() {
		case ItemIdentifier, ItemNumber, ItemDoubleQuotedString, ItemSingleQuotedString:
			// okay, proceed
		default:
			break OUTER
		}
		item := b.NextNonSpace(ctx)

		// Depending on the next item, we have range operator or a literal list
		var child node.Node
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

func (b *Builder) ParseIf(ctx *builderCtx) node.Node{
	ifToken := b.NextNonSpace(ctx)
	if ifToken.Type() != ItemIf {
		b.Unexpected(ctx, "Expected if, got %s", ifToken)
	}

	// parenthesis are optional
	expectCloseParen := false
	if b.PeekNonSpace(ctx).Type() == ItemOpenParen {
		b.NextNonSpace(ctx)
		expectCloseParen = true
	}

	exp := b.ParseExpression(ctx, false)
	ifNode := node.NewIfNode(ifToken.Pos(), exp)

	if expectCloseParen {
		closeParenToken := b.NextNonSpace(ctx)
		if closeParenToken.Type() != ItemCloseParen {
			b.Unexpected(ctx, "Expected close parenthesis, got %s", closeParenToken)
		}
	}

	ctx.CurrentParentNode().Append(ifNode)
	ctx.PushParentNode(ifNode)

	return nil
}

func (b *Builder) ParseElse(ctx *builderCtx) node.Node{
	elseToken := b.NextNonSpace(ctx)
	if elseToken.Type() != ItemElse {
		b.Unexpected(ctx, "Expected else, got %s", elseToken)
	}

	// CurrentParentNode must be "If" in order for "else" to work
	if ctx.CurrentParentNode().Type() != node.If {
		b.Unexpected(ctx, "Found else without if")
	}

	elseNode := node.NewElseNode(elseToken.Pos())
	elseNode.IfNode = ctx.CurrentParentNode()
	ctx.CurrentParentNode().Append(elseNode)
	ctx.PushParentNode(elseNode)

	return nil
}

func (b *Builder) ParseInclude(ctx *builderCtx) node.Node{
	incToken := b.NextNonSpace(ctx)
	if incToken.Type() != ItemInclude {
		b.Unexpected(ctx, "Expected include, got %s", incToken)
	}

	// Next thing must be the name of the included template
	n := b.ParseExpression(ctx, false)
	x := node.NewIncludeNode(incToken.Pos(), n)
	ctx.PushFrame()

	if b.PeekNonSpace(ctx).Type() != ItemWith {
		ctx.PopFrame()
		return x
	}

	b.NextNonSpace(ctx)
	for {
		a := b.ParseAssignment(ctx)
		x.AppendAssignment(a)
		next := b.PeekNonSpace(ctx)
		if next.Type() != ItemComma {
			break
		} else if next.Type() == ItemTagEnd {
			break
		}
		b.NextNonSpace(ctx)
	}
	ctx.PopFrame()

	return x
}

func (b *Builder) ParseGroup(ctx *builderCtx) node.Node{
	openParenToken := b.NextNonSpace(ctx)
	if openParenToken.Type() != ItemOpenParen {
		b.Unexpected(ctx, "Expected '(', got %s", openParenToken)
	}

	n := node.NewGroupNode(openParenToken.Pos())
	n.Child = b.ParseExpression(ctx, false)

	closeParenToken := b.NextNonSpace(ctx)
	if closeParenToken.Type() != ItemCloseParen {
		b.Unexpected(ctx, "Expected ')', got %s", closeParenToken)
	}

	return n
}

func (b *Builder) ParseMacro(ctx *builderCtx) node.Node{
	macroToken := b.NextNonSpace(ctx)
	if macroToken.Type() != ItemMacro {
		b.Unexpected(ctx, "Expected 'MACRO', got %s", macroToken)
	}

	// Parse name, and arguments.
	nameToken := b.NextNonSpace(ctx)
	if nameToken.Type() != ItemIdentifier {
		b.Unexpected(ctx, "Expected identifier, got %s", nameToken)
	}

	idx := ctx.DeclareLocalVar(nameToken.Value())

	macro := node.NewMacroNode(nameToken.Pos(), nameToken.Value())
	macro.LocalVar = node.NewLocalVarNode(nameToken.Pos(), nameToken.Value(), idx)
	ctx.CurrentParentNode().Append(macro)
	ctx.PushParentNode(macro)

	// either a '(' followed by argument list, or BLOCK
	if b.PeekNonSpace(ctx).Type() == ItemOpenParen {
		b.NextNonSpace(ctx) // discard open paren
		// Can't use ParseList() here, because we want a list of only identifiers
		for {
			next := b.NextNonSpace(ctx)
			if next.Type() != ItemIdentifier {
				b.Backup(ctx)
				break
			}

			//      idx := ctx.DeclareLocalVar(next.Value())
			//      macro.AppendArg(node.NewLocalVarNode(next.Pos(), next.Value(), idx))

			next = b.NextNonSpace(ctx)
			if next.Type() != ItemComma {
				b.Backup(ctx)
				break
			}
		}
		if closeParen := b.NextNonSpace(ctx); closeParen.Type() != ItemCloseParen {
			b.Unexpected(ctx, "Expected ')', got %s", closeParen)
		}
	}

	// Then we need a BLOCK
	if block := b.NextNonSpace(ctx); block.Type() != ItemBlock {
		b.Unexpected(ctx, "Expected BLOCK, got %s", block)
	}

	return nil
}
