package node

import "reflect"

const DefaultMaxIterations = 1000

// NodeType is used to distinguish each AST node
type NodeType int

// Node defines the interface for an AST node
type Node interface {
	Type() NodeType
	Copy() Node
	Pos() int
	Visit(chan Node)
}

type Appender interface {
	Node
	Append(Node)
}

const (
	Noop NodeType = iota
	Root
	Text
	Number
	Int
	Float
	If
	Else
	List
	Foreach
	While
	Wrapper
	Include
	Assignment
	LocalVar
	FetchField
	FetchArrayElement
	MethodCall
	FunCall
	Print
	PrintRaw
	FetchSymbol
	Range
	Plus
	Minus
	Mul
	Div
	Equals
	NotEquals
	LT
	GT
	MakeArray
	Group
	Filter
	Macro
	Max
)

// BaseNode is the most basic node with no extra data attached to it
type BaseNode struct {
	NodeType // String() is delegated here
	pos      int
}

type ListNode struct {
	BaseNode
	Nodes []Node
}

type NumberNode struct {
	BaseNode
	Value reflect.Value
}

type TextNode struct {
	BaseNode
	Text []byte
}

type WrapperNode struct {
	*ListNode
	WrapperName string
	// XXX need to make this configurable. currently it's only "content"
	// WrapInto string
	AssignmentNodes []Node
}

type AssignmentNode struct {
	BaseNode
	Assignee   *LocalVarNode
	Expression Node
}

type LocalVarNode struct {
	BaseNode
	Name   string
	Offset int
}

type LoopNode struct {
	*ListNode         // Body of the loop
	Condition    Node //
	MaxIteration int  // Max number of iterations
}

type ForeachNode struct {
	*LoopNode
	IndexVarName string
	IndexVarIdx  int
	List         Node
}

type WhileNode struct {
	*LoopNode
}

type MethodCallNode struct {
	BaseNode
	Invocant   Node
	MethodName string
	Args       *ListNode
}

type FunCallNode struct {
	BaseNode
	Invocant Node
	Args     *ListNode
}

type FetchFieldNode struct {
	BaseNode
	Container Node
	FieldName string
}

type IfNode struct {
	*ListNode
	BooleanExpression Node
}

type ElseNode struct {
	*ListNode
	IfNode Node
}

type UnaryNode struct {
	BaseNode
	Child Node
}

type BinaryNode struct {
	BaseNode
	Left  Node
	Right Node
}

type FilterNode struct {
	*UnaryNode
	Name string
}

type MacroNode struct {
	*ListNode
	Name      string
	LocalVar  *LocalVarNode
	Arguments []*LocalVarNode
}
