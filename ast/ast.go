package ast

import (
  "time"
)

type INode interface {
  // TBD
}

type AST struct {
  root INode
  charset string // utf8
  timestamp time.Time
}

func NewAST () *AST {
  return &AST {
    charset: "utf8",
    timestamp: time.Now(),
    root: &AST {
    },
  }
}

