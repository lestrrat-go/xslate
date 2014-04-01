
package compiler

import (
  "fmt"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

// Compiler is the interface to objects that can convert AST trees to
// actual Xslate Virtual Machine bytecode (see vm.ByteCode)
type Compiler interface {
  Compile(* parser.AST) (*vm.ByteCode, error)
}

type context struct {
  ByteCode *vm.ByteCode
}

// AppendOp creates and appends a new op to the current set of ByteCode
func (ctx *context) AppendOp(o vm.OpType, args ...interface {}) *vm.Op {
  return ctx.ByteCode.AppendOp(o, args...)
}

// BasicCompiler is the default compiler used by Xslate
type BasicCompiler struct {}

// New creates a new BasicCompiler instance
func New() *BasicCompiler {
  return &BasicCompiler {}
}

// Compile satisfies the compiler.Compiler interface. It accepts an AST
// created by parser.Parser, and returns vm.ByteCode or an error
func (c *BasicCompiler) Compile(ast *parser.AST) (*vm.ByteCode, error) {
  ctx := &context {
    ByteCode: vm.NewByteCode(),
  }
  for _, n := range ast.Root.Nodes {
    c.compile(ctx, n)
  }

  // When we're done compiling, always append an END op
  ctx.ByteCode.AppendOp(vm.TXOPEnd)

  return ctx.ByteCode, nil
}

func (c *BasicCompiler) compile(ctx *context, n parser.Node) {
  switch n.Type() {
  case parser.NodeText:
    // XXX probably not true all the time
    ctx.AppendOp(vm.TXOPLiteral, n.(*parser.TextNode).Text)
  case parser.NodeFetchSymbol:
    ctx.AppendOp(vm.TXOPFetchSymbol, n.(*parser.TextNode).Text)
  case parser.NodeFetchField:
    ffnode := n.(*parser.FetchFieldNode)
    c.compile(ctx, ffnode.Container)
    ctx.AppendOp(vm.TXOPFetchFieldSymbol, ffnode.FieldName)
  case parser.NodeFetchArrayElement:
    faenode := n.(*parser.BinaryNode)
    ctx.AppendOp(vm.TXOPPushmark)
    c.compile(ctx, faenode.Right)
    ctx.AppendOp(vm.TXOPPush)
    c.compile(ctx, faenode.Left)
    ctx.AppendOp(vm.TXOPPush)
    ctx.AppendOp(vm.TXOPFetchArrayElement)
    ctx.AppendOp(vm.TXOPPopmark)
  case parser.NodeLocalVar:
    l := n.(*parser.LocalVarNode)
    ctx.AppendOp(vm.TXOPLoadLvar, l.Offset)
  case parser.NodeAssignment:
    c.compile(ctx, n.(*parser.AssignmentNode).Expression)
    ctx.AppendOp(vm.TXOPSaveToLvar, 0) // XXX this 0 must be pre-computed
  case parser.NodePrint:
    c.compile(ctx, n.(*parser.ListNode).Nodes[0])
    ctx.AppendOp(vm.TXOPPrint)
  case parser.NodePrintRaw:
    c.compile(ctx, n.(*parser.ListNode).Nodes[0])
    ctx.AppendOp(vm.TXOPPrintRaw)
  case parser.NodeForeach:
    c.compileForeach(ctx, n.(*parser.ForeachNode))
  case parser.NodeWhile:
    c.compileWhile(ctx, n.(*parser.WhileNode))
  case parser.NodeIf:
    c.compileIf(ctx, n)
  case parser.NodeElse:
    gotoOp := ctx.AppendOp(vm.TXOPGoto, 0)
    pos := ctx.ByteCode.Len()
    for _, child := range n.(*parser.ElseNode).ListNode.Nodes {
      c.compile(ctx, child)
    }
    gotoOp.SetArg(ctx.ByteCode.Len() - pos + 1)
  case parser.NodeMakeArray:
    x := n.(*parser.UnaryNode)
    c.compile(ctx, x.Child)
    ctx.AppendOp(vm.TXOPMakeArray)
  case parser.NodeRange:
    x := n.(*parser.BinaryNode)
    c.compile(ctx, x.Right)
    ctx.AppendOp(vm.TXOPPush)
    c.compile(ctx, x.Left)
    ctx.AppendOp(vm.TXOPMoveToSb)
    ctx.AppendOp(vm.TXOPPop)
    ctx.AppendOp(vm.TXOPRange)
  case parser.NodeInt:
    x := n.(*parser.NumberNode)
    ctx.AppendOp(vm.TXOPLiteral, x.Value.Int())
  case parser.NodeList:
    x := n.(*parser.ListNode)
    for _, v := range x.Nodes {
      c.compile(ctx, v)
      if v.Type() != parser.NodeRange {
        ctx.AppendOp(vm.TXOPPush)
      }
    }
  case parser.NodeFunCall:
    x := n.(*parser.FunCallNode)

    for _, child := range x.Args.Nodes {
      c.compile(ctx, child)
      ctx.AppendOp(vm.TXOPPush)
    }

    c.compile(ctx, x.Invocant)
    ctx.AppendOp(vm.TXOPFunCallOmni)
  case parser.NodeMethodCall:
    x := n.(*parser.MethodCallNode)

    c.compile(ctx, x.Invocant)
    ctx.AppendOp(vm.TXOPPush)
    ctx.AppendOp(vm.TXOPPushmark)
    for _, child := range x.Args.Nodes {
      c.compile(ctx, child)
      ctx.AppendOp(vm.TXOPPush)
    }
    ctx.AppendOp(vm.TXOPMethodCall, x.MethodName)
    ctx.AppendOp(vm.TXOPPopmark)
  case parser.NodeWrapper:
    x := n.(*parser.WrapperNode)
    ctx.AppendOp(vm.TXOPPushOutput)
    ctx.AppendOp(vm.TXOPNewOutput)
    for _, v := range x.ListNode.Nodes {
      c.compile(ctx, v)
    }
    ctx.AppendOp(vm.TXOPPopOutput)

    ctx.AppendOp(vm.TXOPPush)
    // Arguments to include (WITH foo = "bar") need to be evaulated
    // in the OUTER context, but the variables need to be set in the
    // include context
    c.compileAssignmentNodes(ctx, x.AssignmentNodes)
    ctx.AppendOp(vm.TXOPPop)
    ctx.AppendOp(vm.TXOPPushmark)
    ctx.AppendOp(vm.TXOPWrapper, x.WrapperName)
    ctx.AppendOp(vm.TXOPPopmark)
  case parser.NodeInclude:
    x := n.(*parser.IncludeNode)

    c.compile(ctx, x.IncludeTarget)
    ctx.AppendOp(vm.TXOPPush)
    // Arguments to include (WITH foo = "bar") need to be evaulated
    // in the OUTER context, but the variables need to be set in the
    // include context
    c.compileAssignmentNodes(ctx, x.AssignmentNodes)
    ctx.AppendOp(vm.TXOPPop)
    ctx.AppendOp(vm.TXOPPushmark)
    ctx.AppendOp(vm.TXOPInclude)
    ctx.AppendOp(vm.TXOPPopmark)
  case parser.NodeGroup:
    c.compile(ctx, n.(*parser.UnaryNode).Child)
  case parser.NodeEquals, parser.NodeNotEquals, parser.NodeLT, parser.NodeGT:
    x := n.(*parser.BinaryNode)

    c.compileBinaryOperands(ctx, x)
    switch n.Type() {
    case parser.NodeEquals:
      ctx.AppendOp(vm.TXOPEquals)
    case parser.NodeNotEquals:
      ctx.AppendOp(vm.TXOPNotEquals)
    case parser.NodeLT:
      ctx.AppendOp(vm.TXOPLessThan)
    case parser.NodeGT:
      ctx.AppendOp(vm.TXOPGreaterThan)
    default:
      panic("Unknown operator")
    }
  case parser.NodePlus, parser.NodeMinus, parser.NodeMul, parser.NodeDiv:
    x := n.(*parser.BinaryNode)

    c.compileBinaryOperands(ctx, x)
    switch n.Type() {
    case parser.NodePlus:
      ctx.AppendOp(vm.TXOPAdd)
    case parser.NodeMinus:
      ctx.AppendOp(vm.TXOPSub)
    case parser.NodeMul:
      ctx.AppendOp(vm.TXOPMul)
    case parser.NodeDiv:
      ctx.AppendOp(vm.TXOPDiv)
    default:
      panic("Unknown arithmetic")
    }
  case parser.NodeFilter:
    x := n.(*parser.FilterNode)

    c.compile(ctx, x.Child)
    ctx.AppendOp(vm.TXOPFilter, x.Name)
  case parser.NodeMacro:
    x := n.(*parser.MacroNode)
    // The VM is responsible for passing arguments, which do not need
    // to be declared as variables in the template. n.Arguments exists,
    // but it's left untouched

    // This goto effectively forces the VM to "ignore" this block of
    // MACRO definition.
    gotoOp := ctx.AppendOp(vm.TXOPGoto, 0)
    start := ctx.ByteCode.Len()

    // This is the actual "entry point"
    ctx.AppendOp(vm.TXOPPushmark)
    entryPoint := ctx.ByteCode.Len() - 1

    for _, child := range x.Nodes {
      c.compile(ctx, child)
    }
    ctx.AppendOp(vm.TXOPPopmark)
    ctx.AppendOp(vm.TXOPEnd) // This END forces termination
    gotoOp.SetArg(ctx.ByteCode.Len() - start + 1)

    // Now remember about this definition
    ctx.AppendOp(vm.TXOPLiteral, entryPoint)
    ctx.AppendOp(vm.TXOPSaveToLvar, x.LocalVar.Offset)
  default:
    fmt.Printf("Unknown node: %s\n", n.Type())
  }
}

func (c *BasicCompiler) compileIf(ctx *context, n parser.Node) {
  x := n.(*parser.IfNode)
  ctx.AppendOp(vm.TXOPPushmark)
  c.compile(ctx, x.BooleanExpression)
  ifop := ctx.AppendOp(vm.TXOPAnd, 0)
  pos := ctx.ByteCode.Len()

  var elseNode parser.Node
  children := x.ListNode.Nodes
  for _, child := range children {
    if child.Type() == parser.NodeElse {
      elseNode = child
    } else {
      c.compile(ctx, child)
    }
  }

  if elseNode == nil {
    ifop.SetArg(ctx.ByteCode.Len() - pos + 1)
  } else {
    // If we have an else, we need to put this AFTER the goto
    // that's generated by else
    ifop.SetArg(ctx.ByteCode.Len() - pos + 2)
    c.compile(ctx, elseNode)
  }
  ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileBinaryOperands(ctx *context, x *parser.BinaryNode) {
  if x.Right.Type() == parser.NodeGroup {
    // Grouped node
    c.compile(ctx, x.Right)
    ctx.AppendOp(vm.TXOPPush)
    c.compile(ctx, x.Left)
    ctx.AppendOp(vm.TXOPMoveToSb)
    ctx.AppendOp(vm.TXOPPop)
  } else {
    c.compile(ctx, x.Left)
    ctx.AppendOp(vm.TXOPMoveToSb)
    c.compile(ctx, x.Right)
  }
}

func (c *BasicCompiler) compileAssignmentNodes(ctx *context, assignnodes []parser.Node) {
  if len(assignnodes) <= 0 {
    return
  }
  ctx.AppendOp(vm.TXOPPushmark)
  for _, nv := range assignnodes {
    v := nv.(*parser.AssignmentNode)
    ctx.AppendOp(vm.TXOPLiteral, v.Assignee.Name)
    ctx.AppendOp(vm.TXOPPush)
    c.compile(ctx, v.Expression)
    ctx.AppendOp(vm.TXOPPush)
  }
  ctx.AppendOp(vm.TXOPMakeHash)
  ctx.AppendOp(vm.TXOPMoveToSb)
  ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileForeach(ctx *context, x *parser.ForeachNode) {
  ctx.AppendOp(vm.TXOPPushmark)
  ctx.AppendOp(vm.TXOPPushFrame)
  c.compile(ctx, x.List)
  ctx.AppendOp(vm.TXOPForStart, x.IndexVarIdx)
  ctx.AppendOp(vm.TXOPLiteral, x.IndexVarIdx)

  iter := ctx.AppendOp(vm.TXOPForIter, 0)
  pos  := ctx.ByteCode.Len()

  children := x.Nodes
  for _, v := range children {
    c.compile(ctx, v)
  }

  ctx.AppendOp(vm.TXOPGoto, -1 * (ctx.ByteCode.Len() - pos + 2))
  iter.SetArg(ctx.ByteCode.Len() - pos + 1)
  ctx.AppendOp(vm.TXOPPopFrame)
  ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileWhile(ctx *context, x *parser.WhileNode) {
  ctx.AppendOp(vm.TXOPPushmark)
  condPos := ctx.ByteCode.Len() + 1 // w/o 1, it's the pushmark, but we want the next one

  // compile the boolean expression
  c.compile(ctx, x.Condition)

  // we might as well use the equivalent of If here!
  ifop := ctx.AppendOp(vm.TXOPAnd, 0)
  ifPos := ctx.ByteCode.Len()

  children := x.Nodes
  for _, v := range children {
    c.compile(ctx, v)
  }

  // Go back to condPos
  ctx.AppendOp(vm.TXOPGoto, -1 * (ctx.ByteCode.Len() - condPos + 1))
  ifop.SetArg(ctx.ByteCode.Len() - ifPos + 1)
  ctx.AppendOp(vm.TXOPPopmark)
}
