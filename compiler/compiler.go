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
  if n.Type() == parser.NodeText {
    ctx.ByteCode.AppendOp(vm.TXOP_print_raw, n.(*parser.TextNode).Text)
  }
}