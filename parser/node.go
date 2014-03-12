package parser
import (
  "bytes"
  "fmt"
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
}

const (
  NodeNoop NodeType = iota
  NodeRoot
  NodeText
  NodeNumber
  NodeList
  NodeForeach
  NodeWrapper
  NodeAssignment
  NodeLocalVar
  NodeFetchField
  NodeMethodcall
)

func (n NodeType) String() string {
  switch n {
  case NodeNoop:
    return "Noop"
  case NodeRoot:
    return "Root"
  case NodeText:
    return "Text"
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

func NewNode(t NodeType, pos Pos, args ...interface {}) Node {
  switch t {
  case NodeNoop:
    return NewNoopNode()
  case NodeList:
    return NewListNode(pos)
  case NodeText:
    return NewTextNode(pos, args[0].(string))
  default:
    panic("fuck")
  }
  return nil
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
  n.Append(NewLocalVarNode(pos, symbol))
  return n
}

func NewLocalVarNode(pos Pos, symbol string) *TextNode {
  n := NewTextNode(pos, symbol)
  n.NodeType = NodeLocalVar
  return n
}

func NewForeachNode(pos Pos, symbol string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeForeach
  n.Append(NewLocalVarNode(pos, symbol))
  return n
}

func NewMethodcallNode(pos Pos, invocant, method string, args Node) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeMethodcall
  n.Append(NewLocalVarNode(pos, invocant))
  n.Append(NewTextNode(pos, method))
  n.Append(args)
  return n
}

func NewFetchFieldNode(pos Pos, invocant, field string) *ListNode {
  n := NewListNode(pos)
  n.NodeType = NodeFetchField
  n.Append(NewLocalVarNode(pos, invocant))
  n.Append(NewTextNode(pos, field))
  return n
}

func NewRootNode() *ListNode {
  n := NewListNode(0)
  n.NodeType = NodeRoot
  return n
}

func NewNumberNode(pos Pos, number string) *TextNode {
  n := NewTextNode(pos, number)
  n.NodeType = NodeNumber
  return n
}
