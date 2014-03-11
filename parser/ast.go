package parser

import (
  "time"
)

type AST struct {
  Name      string // name of the template
  ParseName string // name of the top-level template during parsing
  Root      *ListNode // root of the tree
  Timestamp time.Time // last-modified date of this template
  text      string
}

/*
func NewAST () *AST {
  return &AST {
    Timestamp: time.Now(),
    Root: NewNodeList(),
  }
}
*/
