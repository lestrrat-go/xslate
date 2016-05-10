package node

import (
	"fmt"
	"reflect"
)

// Type returns the current node type
func (n NodeType) Type() NodeType {
	return n
}

// Pos returns the position of this node in the document
func (n *BaseNode) Pos() int {
	return n.pos
}

func (n *BaseNode) Copy() Node {
	return &BaseNode{n.NodeType, n.pos}
}

func (n *BaseNode) Visit(c chan Node) {
	c <- n
}

// Noop nodes don't need to be distinct
var noop = &BaseNode{Noop, 0}

// NewNoopNode returns a op that does nothing
func NewNoopNode() *BaseNode {
	return noop
}

func (l *ListNode) Visit(c chan Node) {
	c <- l
	for _, child := range l.Nodes {
		child.Visit(c)
	}
}

func NewListNode(pos int) *ListNode {
	return &ListNode{
		BaseNode{List, pos},
		[]Node{},
	}
}

func (l *ListNode) Copy() Node {
	n := NewListNode(l.pos)
	n.Nodes = make([]Node, len(l.Nodes))
	for k, v := range l.Nodes {
		n.Nodes[k] = v.Copy()
	}
	return n
}

func (l *ListNode) Append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

func NewTextNode(pos int, arg string) *TextNode {
	return &TextNode{
		BaseNode{Text, pos},
		[]byte(arg),
	}
}

func (n *TextNode) Copy() Node {
	return NewTextNode(n.pos, string(n.Text))
}

func (n *TextNode) String() string {
	return fmt.Sprintf("%s %q", n.NodeType, n.Text)
}

func (n *TextNode) Visit(c chan Node) {
	c <- n
}

func NewWrapperNode(pos int, template string) *WrapperNode {
	n := &WrapperNode{
		NewListNode(pos),
		template,
		[]Node{},
	}
	n.NodeType = Wrapper
	return n
}

func (n *WrapperNode) AppendAssignment(a Node) {
	n.AssignmentNodes = append(n.AssignmentNodes, a)
}

func (n *WrapperNode) Copy() Node {
	anodes := make([]Node, len(n.AssignmentNodes))
	for i, v := range n.AssignmentNodes {
		anodes[i] = v.Copy()
	}
	return &WrapperNode{
		n.ListNode.Copy().(*ListNode),
		n.WrapperName,
		anodes,
	}
}

func (n *WrapperNode) Visit(c chan Node) {
	c <- n
	for _, v := range n.AssignmentNodes {
		v.Visit(c)
	}
	n.ListNode.Visit(c)
}

func NewAssignmentNode(pos int, symbol string) *AssignmentNode {
	n := &AssignmentNode{
		BaseNode{Assignment, pos},
		NewLocalVarNode(pos, symbol, 0), // TODO
		nil,
	}
	return n
}

func (n *AssignmentNode) Copy() Node {
	x := &AssignmentNode{
		BaseNode{Assignment, n.pos},
		n.Assignee,
		n.Expression,
	}
	return x
}

func (n *AssignmentNode) Visit(c chan Node) {
	c <- n
	c <- n.Assignee
	c <- n.Expression
}

func NewLocalVarNode(pos int, symbol string, idx int) *LocalVarNode {
	n := &LocalVarNode{
		BaseNode{LocalVar, pos},
		symbol,
		idx,
	}
	return n
}

func (n *LocalVarNode) Copy() Node {
	return NewLocalVarNode(n.pos, n.Name, n.Offset)
}

func (n *LocalVarNode) Visit(c chan Node) {
	c <- n
}

func (n *LocalVarNode) String() string {
	return fmt.Sprintf("%s %s (%d)", n.NodeType, n.Name, n.Offset)
}

func NewLoopNode(pos int) *LoopNode {
	maxiter := DefaultMaxIterations
	return &LoopNode{
		ListNode:     NewListNode(pos),
		MaxIteration: maxiter,
	}
}

func NewForeachNode(pos int, symbol string) *ForeachNode {
	n := &ForeachNode{
		IndexVarName: symbol,
		IndexVarIdx:  0,
		List:         nil,
		LoopNode:     NewLoopNode(pos),
	}
	n.NodeType = Foreach
	return n
}

func (n *ForeachNode) Visit(c chan Node) {
	c <- n
	// Skip the list node that we contain
	for _, child := range n.ListNode.Nodes {
		child.Visit(c)
	}
}

func (n *ForeachNode) Copy() Node {
	x := &ForeachNode{
		IndexVarName: n.IndexVarName,
		IndexVarIdx:  n.IndexVarIdx,
		List:         n.List.Copy(),
		LoopNode:     n.LoopNode.Copy().(*LoopNode),
	}
	x.NodeType = Foreach
	return n
}

func (n *ForeachNode) String() string {
	return fmt.Sprintf("%s %s (%d)", n.NodeType, n.IndexVarName, n.IndexVarIdx)
}

func NewWhileNode(pos int, n Node) *WhileNode {
	x := &WhileNode{
		LoopNode: NewLoopNode(pos),
	}
	x.Condition = n
	x.NodeType = While
	return x
}

func (n *WhileNode) Copy() Node {
	return &WhileNode{
		LoopNode: n.LoopNode.Copy().(*LoopNode),
	}
}

func (n *WhileNode) Visit(c chan Node) {
	c <- n
	n.Condition.Visit(c)
	n.ListNode.Visit(c)
}

func NewMethodCallNode(pos int, invocant Node, method string, args *ListNode) *MethodCallNode {
	return &MethodCallNode{
		BaseNode{MethodCall, pos},
		invocant,
		method,
		args,
	}
}

func (n *MethodCallNode) Copy() Node {
	return NewMethodCallNode(n.pos, n.Invocant, n.MethodName, n.Args)
}

func (n *MethodCallNode) Visit(c chan Node) {
	c <- n
	n.Invocant.Visit(c)
	n.Args.Visit(c)
}

func NewFunCallNode(pos int, invocant Node, args *ListNode) *FunCallNode {
	return &FunCallNode{
		BaseNode{FunCall, pos},
		invocant,
		args,
	}
}

func (n *FunCallNode) Copy() Node {
	return NewFunCallNode(n.pos, n.Invocant, n.Args)
}

func (n *FunCallNode) Visit(c chan Node) {
	c <- n
	n.Invocant.Visit(c)
	n.Args.Visit(c)
}

func NewFetchFieldNode(pos int, container Node, field string) *FetchFieldNode {
	n := &FetchFieldNode{
		BaseNode{FetchField, pos},
		container,
		field,
	}
	return n
}

func (n *FetchFieldNode) Copy() Node {
	return &FetchFieldNode{
		BaseNode{FetchField, n.pos},
		n.Container.Copy(),
		n.FieldName,
	}
}

func (n *FetchFieldNode) Visit(c chan Node) {
	c <- n
	n.Container.Visit(c)
}

func NewRootNode() *ListNode {
	n := NewListNode(0)
	n.NodeType = Root
	return n
}

func NewNumberNode(pos int, num reflect.Value) *NumberNode {
	return &NumberNode{
		BaseNode{Number, pos},
		num,
	}
}

func (n *NumberNode) Copy() Node {
	x := NewNumberNode(n.pos, n.Value)
	x.NodeType = n.NodeType
	return x
}

func (n *NumberNode) Visit(c chan Node) {
	c <- n
}

func NewIntNode(pos int, v int64) *NumberNode {
	n := NewNumberNode(pos, reflect.ValueOf(v))
	n.NodeType = Int
	return n
}

func NewFloatNode(pos int, v float64) *NumberNode {
	n := NewNumberNode(pos, reflect.ValueOf(v))
	n.NodeType = Float
	return n
}

func NewPrintNode(pos int, arg Node) *ListNode {
	n := NewListNode(pos)
	n.NodeType = Print
	n.Append(arg)
	return n
}

func NewPrintRawNode(pos int) *ListNode {
	n := NewListNode(pos)
	n.NodeType = PrintRaw
	return n
}

func NewFetchSymbolNode(pos int, symbol string) *TextNode {
	n := NewTextNode(pos, symbol)
	n.NodeType = FetchSymbol
	return n
}

func NewIfNode(pos int, exp Node) *IfNode {
	n := &IfNode{
		NewListNode(pos),
		exp,
	}
	n.NodeType = If
	return n
}

func (n *IfNode) Copy() Node {
	x := &IfNode{
		n.ListNode.Copy().(*ListNode),
		nil,
	}
	if e := n.BooleanExpression; e != nil {
		x.BooleanExpression = e.Copy()
	}

	x.ListNode = n.ListNode.Copy().(*ListNode)

	return x
}

func (n *IfNode) Visit(c chan Node) {
	c <- n
	c <- n.BooleanExpression
	for _, child := range n.ListNode.Nodes {
		c <- child
	}
}

func NewElseNode(pos int) *ElseNode {
	n := &ElseNode{
		NewListNode(pos),
		nil,
	}
	n.NodeType = Else
	return n
}

func NewRangeNode(pos int, start, end Node) *BinaryNode {
	return &BinaryNode{
		BaseNode{Range, pos},
		start,
		end,
	}
}

func (n *UnaryNode) Visit(c chan Node) {
	c <- n
	n.Child.Visit(c)
}

func (n *UnaryNode) Copy() Node {
	return &UnaryNode{
		BaseNode{n.NodeType, n.pos},
		n.Child.Copy(),
	}
}

func NewMakeArrayNode(pos int, child Node) *UnaryNode {
	return &UnaryNode{
		BaseNode{MakeArray, pos},
		child,
	}
}

type IncludeNode struct {
	BaseNode
	IncludeTarget   Node
	AssignmentNodes []Node
}

func NewIncludeNode(pos int, include Node) *IncludeNode {
	return &IncludeNode{
		BaseNode{Include, pos},
		include,
		[]Node{},
	}
}

func (n *IncludeNode) AppendAssignment(a Node) {
	n.AssignmentNodes = append(n.AssignmentNodes, a)
}

func (n *IncludeNode) Copy() Node {
	return NewIncludeNode(n.pos, n.IncludeTarget)
}

func (n *IncludeNode) Visit(c chan Node) {
	c <- n
	c <- n.IncludeTarget
}

func NewPlusNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{Plus, pos},
		nil,
		nil,
	}
}

func NewMinusNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{Minus, pos},
		nil,
		nil,
	}
}

func NewMulNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{Mul, pos},
		nil,
		nil,
	}
}

func NewDivNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{Div, pos},
		nil,
		nil,
	}
}

func NewEqualsNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{Equals, pos},
		nil,
		nil,
	}
}

func NewNotEqualsNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{NotEquals, pos},
		nil,
		nil,
	}
}

func NewLTNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{LT, pos},
		nil,
		nil,
	}
}

func NewGTNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{GT, pos},
		nil,
		nil,
	}
}

func (n *BinaryNode) Copy() Node {
	return &BinaryNode{
		BaseNode{n.NodeType, n.pos},
		n.Left.Copy(),
		n.Right.Copy(),
	}
}

func (n *BinaryNode) Visit(c chan Node) {
	c <- n
	n.Left.Visit(c)
	n.Right.Visit(c)
}

func NewGroupNode(pos int) *UnaryNode {
	return &UnaryNode{
		BaseNode{Group, pos},
		nil,
	}
}

func NewFilterNode(pos int, name string, child Node) *FilterNode {
	return &FilterNode{
		&UnaryNode{
			BaseNode{Filter, pos},
			child,
		},
		name,
	}
}

func (n *FilterNode) Copy() Node {
	return &FilterNode{
		&UnaryNode{
			BaseNode{Filter, n.pos},
			n.Child.Copy(),
		},
		n.Name,
	}
}

func (n *FilterNode) Visit(c chan Node) {
	c <- n
	n.UnaryNode.Visit(c)
}

func NewFetchArrayElementNode(pos int) *BinaryNode {
	return &BinaryNode{
		BaseNode{FetchArrayElement, pos},
		nil,
		nil,
	}
}

func NewMacroNode(pos int, name string) *MacroNode {
	n := &MacroNode{
		NewListNode(pos),
		name,
		nil,
		[]*LocalVarNode{},
	}
	n.NodeType = Macro
	return n
}

func (n *MacroNode) AppendArg(arg *LocalVarNode) {
	n.Arguments = append(n.Arguments, arg)
}
