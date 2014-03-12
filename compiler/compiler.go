package compiler

import (
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

type Compiler interface {
  Compile(* parser.AST) *vm.ByteCode
}

type CompilerCtx struct {
  ByteCode *vm.ByteCode
}

type BasicCompiler struct {}

func New() *BasicCompiler {
  return &BasicCompiler {}
}

func (c *BasicCompiler) Compile(ast *parser.AST) (*vm.ByteCode, error) {
  ctx := &CompilerCtx {
    ByteCode: &vm.ByteCode {},
  }
  for _, n := range ast.Root.Nodes {
    c.compile(ctx, n)
  }

  // When we're done compiling, always append an END op
  ctx.ByteCode.AppendOp(vm.TXOP_end)

  return ctx.ByteCode, nil
}

func (c *BasicCompiler) compile(ctx *CompilerCtx, n parser.Node) {
  switch n.Type() {
  case parser.NodeText:
    // XXX probably not true all the time
    ctx.ByteCode.AppendOp(vm.TXOP_literal, n.(*parser.TextNode).Text)
    ctx.ByteCode.AppendOp(vm.TXOP_print_raw)
  case parser.NodeLocalVar:
    ctx.ByteCode.AppendOp(vm.TXOP_load_lvar, n.(*parser.TextNode).Text)
  case parser.NodePrint:
    c.compile(ctx, n.(*parser.ListNode).Nodes[0])
    ctx.ByteCode.AppendOp(vm.TXOP_print)
  }
}