package parser

import (
	"bytes"
	"fmt"
	"time"

	"github.com/lestrrat/go-xslate/node"
)

// AST is represents the syntax tree for an Xslate template
type AST struct {
	Name      string         // name of the template
	ParseName string         // name of the top-level template during parsing
	Root      *node.ListNode // root of the tree
	Timestamp time.Time      // last-modified date of this template
	text      string
}

// Visit returns a channel which you can receive Node structs in order that
// that they would be processed
func (ast *AST) Visit() <-chan node.Node {
	c := make(chan node.Node)
	go func() {
		ast.Root.Visit(c)
		close(c)
	}()
	return c
}

// String returns the textual representation of this AST
func (ast *AST) String() string {
	buf := &bytes.Buffer{}
	c := ast.Visit()
	k := 0
	for v := range c {
		k++
		fmt.Fprintf(buf, "%03d. %s\n", k, v)
	}
	return buf.String()
}
