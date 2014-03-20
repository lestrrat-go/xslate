package parser

import (
  "bytes"
  "fmt"
  "time"
)

type AST struct {
  Name      string // name of the template
  ParseName string // name of the top-level template during parsing
  Root      *ListNode // root of the tree
  Timestamp time.Time // last-modified date of this template
  text      string
}

func (ast *AST) Visit() <-chan Node {
  c := make(chan Node)
  go func () {
    ast.Root.Visit(c)
    close(c)
  }()
  return c
}

func (ast *AST) String() string {
  buf := &bytes.Buffer {}
  c := ast.Visit()
  k := 0
  for v := range c {
    k++
    fmt.Fprintf(buf, "%03d. %s\n", k, v)
  }
  return buf.String()
}
