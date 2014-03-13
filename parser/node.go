package parser
import (
  "bytes"
  "fmt"
  "reflect"
)

var nodeTextFormat = "%s"

type NodeType int
func (n NodeType) Type() NodeType {
  return n
}

type Node interface {
  Type() NodeType
  String() string
  Copy() Node
  Position() Pos
  Visit(chan Node)
}

const (
  NodeNoop NodeType = iota
  NodeRoot
  NodeText
  NodeNumber
  NodeInt
  NodeFloat
  NodeList
  NodeForeach
  NodeWrapper
  NodeAssignment
  NodeLocalVar
  NodeFetchField
  NodeMethodcall
  NodePrint
  NodePrintRaw
  NodeFetchSymbol
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
  case NodeWrapper:
    return "Wrapper"
  case NodeAssignment:
    return "Assignment"
  case NodeLocalVar:
    return "LocalVar"
  case NodeFetchField:
    return "FetchField"
  case NodeMethodcall:
    return "Methodcall"
  case NodePrint:
    return "Print"
  case NodePrintRaw:
    return "PrintRaw"
  case NodeFetchSymbol:
    return "FetchSymbol"
  default:
    return "Unknown Node"
  }
}

type Pos int

func (p Pos) Position() Pos {
  return p
}

type NoopNode struct {
  NodeType
  Pos
}

type ListNode struct {
  NodeType
  Pos
  Nodes []Node
}

type TextNode struct {
  NodeType
  Pos
  Text []byte
}

type NumberNode struct {
  NodeType
  Pos
  Value reflect.Value
}

func (l *ListNode) Visit(c chan Node) {
  c <- l
  for _, child := range l.Nodes {
    child.Visit(c)
  }
}

func (t *TextNode) Visit(c chan Node) {
  c <- t
}

var noop = &NoopNode {NodeType: NodeNoop}
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

func NewListNode(pos Pos) *ListNode {
  return &ListNode {NodeType: NodeList, Pos: pos, Nodes: []Node {}}
}

func (l ListNode) Copy() Node {
  n := NewListNode(l.Pos)
  n.Nodes = make([]Node, len(l.Nodes))
  copy(n.Nodes, l.Nodes)
  return n
}

func (l *ListNode) String() string {
  b := &bytes.Buffer {}
  for _, n := range l.Nodes {
    fmt.Fprint(b, n)
  }
  return b.String()
}

func (l *ListNode) Append(n Node) {
  l.Nodes = append(l.Nodes, n)
}

func NewTextNode(pos Pos, arg string) *TextNode {
  return &TextNode {NodeType: NodeText, Pos: pos, Text: []byte(arg)}
}

func (n *TextNode) Copy() Node {
  return NewTextNode(n.Pos, string(n.Text))
}

func (n *TextNode) String() string {
  return fmt.Sprintf(nodeTextFormat, n.Text)
}

func NewWrapperNode(pos Pos, template string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeWrapper
  n.Append(NewTextNode(pos, template))
  return n
}

func NewAssignmentNode(pos Pos, symbol string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeAssignment
  n.Append(NewLocalVarNode(pos, symbol, 0)) // TODO
  return n
}

func NewLocalVarNode(pos Pos, symbol string, idx int) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeLocalVar
  n.Append(NewTextNode(pos, symbol))
  n.Append(NewIntNode(pos, int64(idx)))
  return n
}

func NewForeachNode(pos Pos, symbol string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeForeach
  n.Append(NewLocalVarNode(pos, symbol, 0)) // TODO
  return n
}

func NewMethodcallNode(pos Pos, invocant, method string, args Node) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeMethodcall
  n.Append(NewLocalVarNode(pos, invocant, 0)) // TODO
  n.Append(NewTextNode(pos, method))
  n.Append(args)
  return n
}

func NewFetchFieldNode(pos Pos, invocant, field string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeFetchField
  n.Append(NewLocalVarNode(pos, invocant, 0)) // TODO
  n.Append(NewTextNode(pos, field))
  return n
}

func NewRootNode() *ListNode {
  n := NewListNode(0)
  n.NodeType = NodeRoot
  return n
}

func NewNumberNode(pos Pos, num reflect.Value) *NumberNode {
  return &NumberNode {NodeType: NodeNumber, Pos: pos, Value: num}
}

func (n *NumberNode) Copy() Node {
  x := NewNumberNode(n.Pos, n.Value)
  x.NodeType = n.NodeType
  return x
}

func (n *NumberNode) Visit(c chan Node) {
  c <- n
}

func NewIntNode(pos Pos, v int64) *NumberNode {
  n := NewNumberNode(pos, reflect.ValueOf(v))
  n.NodeType = NodeInt
  return n
}

func NewFloatNode(pos Pos, v float64) *NumberNode {
  n := NewNumberNode(pos, reflect.ValueOf(v))
  n.NodeType = NodeFloat
  return n
}

func NewPrintNode(pos Pos, arg Node) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodePrint
  n.Append(arg)
  return n
}

func NewPrintRawNode(pos Pos) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodePrintRaw
  return n
}

func NewFetchSymbolNode(pos Pos, symbol string) *TextNode {
  n := NewTextNode(pos, symbol)
  n.NodeType = NodeFetchSymbol
  return n
}
