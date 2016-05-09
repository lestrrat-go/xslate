package parser

import (
	"bytes"
	"fmt"

	"github.com/lestrrat/go-xslate/node"
)

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
