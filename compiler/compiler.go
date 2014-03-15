package compiler

import (
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

type Compiler interface {
  Compile(* parser.AST) (*vm.ByteCode, error)
}

type CompilerCtx struct {
  ByteCode *vm.ByteCode
}

func (ctx *CompilerCtx) AppendOp(o vm.OpType, args ...interface {}) *vm.Op {
  return ctx.ByteCode.AppendOp(o, args...)
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
    ctx.AppendOp(vm.TXOP_literal, n.(*parser.TextNode).Text)
  case parser.NodeFetchSymbol:
    ctx.AppendOp(vm.TXOP_fetch_s, n.(*parser.TextNode).Text)
  case parser.NodeFetchField:
    ffnode := n.(*parser.FetchFieldNode)
    c.compile(ctx, ffnode.Container)
    ctx.AppendOp(vm.TXOP_fetch_field_s, ffnode.FieldName)
  case parser.NodeLocalVar:
    l := n.(*parser.LocalVarNode)
    ctx.AppendOp(vm.TXOP_load_lvar, l.Offset)
  case parser.NodeAssignment:
    c.compile(ctx, n.(*parser.AssignmentNode).Expression)
    ctx.AppendOp(vm.TXOP_save_to_lvar, 0) // XXX this 0 must be pre-computed
  case parser.NodePrint:
    c.compile(ctx, n.(*parser.ListNode).Nodes[0])
    ctx.AppendOp(vm.TXOP_print)
  case parser.NodePrintRaw:
    c.compile(ctx, n.(*parser.ListNode).Nodes[0])
    ctx.AppendOp(vm.TXOP_print_raw)
  case parser.NodeForeach:
    ctx.AppendOp(vm.TXOP_pushmark)
    c.compile(ctx, n.(*parser.ForeachNode).List)
    ctx.AppendOp(vm.TXOP_for_start, 0)
    ctx.AppendOp(vm.TXOP_literal, 0)
    iter := ctx.AppendOp(vm.TXOP_for_iter, 0)
    pos  := ctx.ByteCode.Len()

    children := n.(*parser.ForeachNode).Nodes
    for _, v := range children {
      c.compile(ctx, v)
    }

    ctx.AppendOp(vm.TXOP_goto, -1 * (ctx.ByteCode.Len() - pos + 2))
    iter.SetArg(ctx.ByteCode.Len() - pos + 1)
    ctx.AppendOp(vm.TXOP_popmark)
  }
}