package parser

import (
	"strings"
	"testing"
)

func TestNode_String(t *testing.T) {
	for i := 0; i < int(NodeMax); i++ {
		nt := NodeType(i)
		if strings.HasPrefix(nt.String(), "Unknown") {
			t.Errorf("%#v does not have String() implemented", nt)
		}
	}
}

func TestTextNode(t *testing.T) {
	n := NewTextNode(0, "foo")
	c := make(chan Node)
	go func() {
		n.Visit(c)
		close(c)
	}()
	for v := range c {
		if _, ok := v.(*TextNode); !ok {
			t.Errorf("expected TextNode, got %v", v)
		}
	}
}
