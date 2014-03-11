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
  NodeText NodeType = iota
  NodeList
)

type Pos int

func (p Pos) Position() Pos {
  return p
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
  case NodeList:
    return NewListNode(pos)
  case NodeText:
    return NewTextNode(pos, args[0].(string))
  default:
    panic("fuck")
  }
  return nil
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
