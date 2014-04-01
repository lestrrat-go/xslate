package parser
import (
  "bytes"
  "fmt"
  "reflect"
)

// NodeType is used to distinguish each AST node
type NodeType int

// Type returns the current node type
func (n NodeType) Type() NodeType {
  return n
}

// Node defines the interface for an AST node
type Node interface {
  Type() NodeType
  String() string
  Copy() Node
  Pos() int
  Visit(chan Node)
}

type NodeAppender interface {
  Node
  Append(Node)
}

const (
  NodeNoop NodeType = iota
  NodeRoot
  NodeText
  NodeNumber
  NodeInt
  NodeFloat
  NodeIf
  NodeElse
  NodeList
  NodeForeach
  NodeWhile
  NodeWrapper
  NodeInclude
  NodeAssignment
  NodeLocalVar
  NodeFetchField
  NodeFetchArrayElement
  NodeMethodCall
  NodeFunCall
  NodePrint
  NodePrintRaw
  NodeFetchSymbol
  NodeRange
  NodePlus
  NodeMinus
  NodeMul
  NodeDiv
  NodeEquals
  NodeNotEquals
  NodeLT
  NodeGT
  NodeMakeArray
  NodeGroup
  NodeFilter
  NodeMacro
  NodeMax
)

func (n NodeType) String() string {
  switch n {
  case NodeNoop:
    return "Noop"
  case NodeRoot:
    return "Root"
  case NodeText:
    return "Text"
  case NodeNumber:
    return "Number"
  case NodeInt:
    return "Int"
  case NodeFloat:
    return "Float"
  case NodeList:
    return "List"
  case NodeForeach:
    return "Foreach"
  case NodeWhile:
    return "While"
  case NodeWrapper:
    return "Wrapper"
  case NodeInclude:
    return "Include"
  case NodeAssignment:
    return "Assignment"
  case NodeLocalVar:
    return "LocalVar"
  case NodeFetchField:
    return "FetchField"
  case NodeFetchArrayElement:
    return "FetchArrayElement"
  case NodeMethodCall:
    return "MethodCall"
  case NodeFunCall:
    return "FunCall"
  case NodePrint:
    return "Print"
  case NodePrintRaw:
    return "PrintRaw"
  case NodeFetchSymbol:
    return "FetchSymbol"
  case NodeIf:
    return "If"
  case NodeElse:
    return "Else"
  case NodeRange:
    return "Range"
  case NodeMakeArray:
    return "MakeArray"
  case NodePlus:
    return "Plus"
  case NodeMinus:
    return "Minus"
  case NodeMul:
    return "Multiply"
  case NodeDiv:
    return "Divide"
  case NodeEquals:
    return "Equals"
  case NodeNotEquals:
    return "NotEquals"
  case NodeLT:
    return "LessThan"
  case NodeGT:
    return "GreaterThan"
  case NodeGroup:
    return "Group"
  case NodeFilter:
    return "Filter"
  case NodeMacro:
    return "Macro"
  default:
    return "Unknown Node"
  }
}

type BaseNode struct {
  NodeType
  pos int
}

func (n *BaseNode) Pos() int {
  return n.pos
}

type NoopNode struct {
  BaseNode
}

type ListNode struct {
  BaseNode
  Nodes []Node
}

type TextNode struct {
  BaseNode
  Text []byte
}

type NumberNode struct {
  BaseNode
  Value reflect.Value
}

func (l *ListNode) Visit(c chan Node) {
  c <- l
  for _, child := range l.Nodes {
    child.Visit(c)
  }
}

func (n *TextNode) Visit(c chan Node) {
  c <- n
}

var noop = &NoopNode { BaseNode { NodeType: NodeNoop, pos: 0 } }
func NewNoopNode() *NoopNode {
  return noop
}

func (n NoopNode) Copy() Node {
  return noop
}

func (n *NoopNode) String() string {
  return "noop"
}

func (n *NoopNode) Visit(chan Node) {
  // ignore
}

func NewListNode(pos int) *ListNode {
  return &ListNode {
    BaseNode { NodeList, pos },
    []Node {},
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
  return &TextNode {
    BaseNode { NodeText, pos },
    []byte(arg),
  }
}

func (n *TextNode) Copy() Node {
  return NewTextNode(n.pos, string(n.Text))
}

func (n *TextNode) String() string {
  return fmt.Sprintf("%s %q", n.NodeType, n.Text)
}

type WrapperNode struct {
  *ListNode
  WrapperName string
  // XXX need to make this configurable. currently it's only "content"
  // WrapInto string
  AssignmentNodes []Node
}

func NewWrapperNode(pos int, template string) *WrapperNode {
  n := &WrapperNode {
    NewListNode(pos),
    template,
    []Node {},
  }
  n.NodeType = NodeWrapper
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
  return &WrapperNode {
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

type AssignmentNode struct {
  BaseNode
  Assignee *LocalVarNode
  Expression Node
}

func NewAssignmentNode(pos int, symbol string) *AssignmentNode {
  n := &AssignmentNode {
    BaseNode { NodeAssignment, pos },
    NewLocalVarNode(pos, symbol, 0), // TODO
    nil,
  }
  return n
}

func (n *AssignmentNode) Copy() Node {
  x := &AssignmentNode {
    BaseNode { NodeAssignment, n.pos },
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

func (n *AssignmentNode) String() string {
  return n.NodeType.String()
}

type LocalVarNode struct {
  BaseNode
  Name string
  Offset int
}

func NewLocalVarNode(pos int, symbol string, idx int) *LocalVarNode {
  n := &LocalVarNode {
    BaseNode { NodeLocalVar, pos },
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

type ForeachNode struct {
  *ListNode
  IndexVarName  string
  IndexVarIdx   int
  List          Node
}

func NewForeachNode(pos int, symbol string) *ForeachNode {
  n := &ForeachNode {
    ListNode: NewListNode(pos),
    IndexVarName: symbol,
    IndexVarIdx: 0,
    List: nil,
  }
  n.NodeType = NodeForeach
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
  x := &ForeachNode {
    ListNode: NewListNode(n.pos),
    IndexVarName: n.IndexVarName,
    IndexVarIdx: n.IndexVarIdx,
    List: n.List.Copy(),
  }
  x.NodeType = NodeForeach
  return n
}

func (n *ForeachNode) String() string {
  b := &bytes.Buffer {}
  fmt.Fprintf(b, "%s %s (%d)", n.NodeType, n.IndexVarName, n.IndexVarIdx)
  return b.String()
}

type WhileNode struct {
  *ListNode
  Condition Node
}

func NewWhileNode(pos int, n Node) *WhileNode {
  x := &WhileNode {
    NewListNode(pos),
    n,
  }
  x.NodeType = NodeWhile
  return x
}

func (n *WhileNode) Copy() Node {
  return &WhileNode {
    n.ListNode.Copy().(*ListNode),
    n.Condition.Copy(),
  }
}

func (n *WhileNode) Visit(c chan Node) {
  c <- n
  n.Condition.Visit(c)
  n.ListNode.Visit(c)
}

type MethodCallNode struct {
  BaseNode
  Invocant Node
  MethodName string
  Args *ListNode
}

func NewMethodCallNode(pos int, invocant Node, method string, args *ListNode) *MethodCallNode {
  return &MethodCallNode {
    BaseNode { NodeMethodCall, pos },
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

type FunCallNode struct {
  BaseNode
  Invocant Node
  Args *ListNode
}

func NewFunCallNode(pos int, invocant Node, args *ListNode) *FunCallNode {
  return &FunCallNode {
    BaseNode { NodeFunCall, pos },
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

type FetchFieldNode struct {
  BaseNode
  Container Node
  FieldName string
}

func NewFetchFieldNode(pos int, container Node, field string) *FetchFieldNode {
  n := &FetchFieldNode {
    BaseNode { NodeFetchField, pos },
    container,
    field,
  }
  return n
}

func (n *FetchFieldNode) Copy() Node {
  return &FetchFieldNode {
    BaseNode { NodeFetchField, n.pos },
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
  n.NodeType = NodeRoot
  return n
}

func NewNumberNode(pos int, num reflect.Value) *NumberNode {
  return &NumberNode {
    BaseNode { NodeNumber, pos },
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
  n.NodeType = NodeInt
  return n
}

func NewFloatNode(pos int, v float64) *NumberNode {
  n := NewNumberNode(pos, reflect.ValueOf(v))
  n.NodeType = NodeFloat
  return n
}

func NewPrintNode(pos int, arg Node) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodePrint
  n.Append(arg)
  return n
}

func NewPrintRawNode(pos int) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodePrintRaw
  return n
}

func NewFetchSymbolNode(pos int, symbol string) *TextNode {
  n := NewTextNode(pos, symbol)
  n.NodeType = NodeFetchSymbol
  return n
}

type IfNode struct {
  *ListNode
  BooleanExpression Node
}

func NewIfNode(pos int, exp Node) *IfNode {
  n := &IfNode {
    NewListNode(pos),
    exp,
  }
  n.NodeType = NodeIf
  return n
}

func (n *IfNode) Copy() Node {
  x := &IfNode {
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

type ElseNode struct {
  *ListNode
  IfNode Node
}

func NewElseNode(pos int) *ElseNode {
  n := &ElseNode {
    NewListNode(pos),
    nil,
  }
  n.NodeType = NodeElse
  return n
}

func NewRangeNode(pos int, start, end Node) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeRange, pos },
    start,
    end,
  }
}

type UnaryNode struct {
  BaseNode
  Child Node
}

func (n *UnaryNode) Visit(c chan Node) {
  c <- n
  n.Child.Visit(c)
}

func (n *UnaryNode) Copy() Node {
  return &UnaryNode {
    BaseNode { n.NodeType, n.pos },
    n.Child.Copy(),
  }
}

func NewMakeArrayNode(pos int, child Node) *UnaryNode {
  return &UnaryNode {
    BaseNode { NodeMakeArray, pos },
    child,
  }
}


type IncludeNode struct {
  BaseNode
  IncludeTarget Node
  AssignmentNodes []Node
}

func NewIncludeNode(pos int, include Node) *IncludeNode {
  return &IncludeNode {
    BaseNode { NodeInclude, pos },
    include,
    []Node {},
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

type BinaryNode struct {
  BaseNode
  Left Node
  Right Node
}

func NewPlusNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodePlus, pos },
    nil,
    nil,
  }
}

func NewMinusNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeMinus, pos },
    nil,
    nil,
  }
}

func NewMulNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeMul, pos },
    nil,
    nil,
  }
}

func NewDivNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeDiv, pos },
    nil,
    nil,
  }
}

func NewEqualsNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeEquals, pos },
    nil,
    nil,
  }
}

func NewNotEqualsNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeNotEquals, pos },
    nil,
    nil,
  }
}

func NewLTNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeLT, pos },
    nil,
    nil,
  }
}

func NewGTNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeGT, pos },
    nil,
    nil,
  }
}

func (n *BinaryNode) Copy() Node {
  return &BinaryNode {
    BaseNode { n.NodeType, n.pos },
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
  return &UnaryNode {
    BaseNode { NodeGroup, pos },
    nil,
  }
}

type FilterNode struct {
  *UnaryNode
  Name string
}

func NewFilterNode(pos int, name string, child Node) *FilterNode {
  return &FilterNode {
    &UnaryNode {
      BaseNode { NodeFilter, pos },
      child,
    },
    name,
  }
}

func (n *FilterNode) Copy() Node {
  return &FilterNode {
    &UnaryNode {
      BaseNode { NodeFilter, n.pos },
      n.Child.Copy(),
    },
    n.Name,
  }
}

func (n *FilterNode) Visit(c chan Node) {
  c <-n
  n.UnaryNode.Visit(c)
}

func NewFetchArrayElementNode(pos int) *BinaryNode {
  return &BinaryNode {
    BaseNode { NodeFetchArrayElement, pos },
    nil,
    nil,
  }
}

type MacroNode struct {
  *ListNode
  Name string
  LocalVar *LocalVarNode
  Arguments []*LocalVarNode
}

func NewMacroNode(pos int, name string) *MacroNode {
  n := &MacroNode {
    NewListNode(pos),
    name,
    nil,
    []*LocalVarNode {},
  }
  n.NodeType = NodeMacro
  return n
}

func (n *MacroNode) AppendArg(arg *LocalVarNode) {
  n.Arguments = append(n.Arguments, arg)
}
