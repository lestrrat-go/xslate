package compiler

import (
	"fmt"

	"github.com/lestrrat/go-xslate/node"
	"github.com/lestrrat/go-xslate/parser"
	"github.com/lestrrat/go-xslate/vm"
)

// Compiler is the interface to objects that can convert AST trees to
// actual Xslate Virtual Machine bytecode (see vm.ByteCode)
type Compiler interface {
	Compile(*parser.AST) (*vm.ByteCode, error)
}

type context struct {
	ByteCode *vm.ByteCode
}

// AppendOp creates and appends a new op to the current set of ByteCode
func (ctx *context) AppendOp(o vm.OpType, args ...interface{}) *vm.Op {
	return ctx.ByteCode.AppendOp(o, args...)
}

// BasicCompiler is the default compiler used by Xslate
type BasicCompiler struct{}

// New creates a new BasicCompiler instance
func New() *BasicCompiler {
	return &BasicCompiler{}
}

// Compile satisfies the compiler.Compiler interface. It accepts an AST
// created by parser.Parser, and returns vm.ByteCode or an error
func (c *BasicCompiler) Compile(ast *parser.AST) (*vm.ByteCode, error) {
	ctx := &context{
		ByteCode: vm.NewByteCode(),
	}
	for _, n := range ast.Root.Nodes {
		c.compile(ctx, n)
	}

	// When we're done compiling, always append an END op
	ctx.ByteCode.AppendOp(vm.TXOPEnd)

	opt := &NaiveOptimizer{}
	opt.Optimize(ctx.ByteCode)

	ctx.ByteCode.Name = ast.Name
	return ctx.ByteCode, nil
}

func (c *BasicCompiler) compile(ctx *context, n node.Node) {
	switch n.Type() {
	case node.NodeText:
		// XXX probably not true all the time
		ctx.AppendOp(vm.TXOPLiteral, n.(*node.TextNode).Text)
	case node.NodeFetchSymbol:
		ctx.AppendOp(vm.TXOPFetchSymbol, n.(*node.TextNode).Text)
	case node.NodeFetchField:
		ffnode := n.(*node.FetchFieldNode)
		c.compile(ctx, ffnode.Container)
		ctx.AppendOp(vm.TXOPFetchFieldSymbol, ffnode.FieldName)
	case node.NodeFetchArrayElement:
		faenode := n.(*node.BinaryNode)
		ctx.AppendOp(vm.TXOPPushmark)
		c.compile(ctx, faenode.Right)
		ctx.AppendOp(vm.TXOPPush)
		c.compile(ctx, faenode.Left)
		ctx.AppendOp(vm.TXOPPush)
		ctx.AppendOp(vm.TXOPFetchArrayElement)
		ctx.AppendOp(vm.TXOPPopmark)
	case node.NodeLocalVar:
		l := n.(*node.LocalVarNode)
		ctx.AppendOp(vm.TXOPLoadLvar, l.Offset)
	case node.NodeAssignment:
		c.compile(ctx, n.(*node.AssignmentNode).Expression)
		ctx.AppendOp(vm.TXOPSaveToLvar, 0) // XXX this 0 must be pre-computed
	case node.NodePrint:
		c.compile(ctx, n.(*node.ListNode).Nodes[0])
		ctx.AppendOp(vm.TXOPPrint)
	case node.NodePrintRaw:
		c.compile(ctx, n.(*node.ListNode).Nodes[0])
		ctx.AppendOp(vm.TXOPPrintRaw)
	case node.NodeForeach:
		c.compileForeach(ctx, n.(*node.ForeachNode))
	case node.NodeWhile:
		c.compileWhile(ctx, n.(*node.WhileNode))
	case node.NodeIf:
		c.compileIf(ctx, n)
	case node.NodeElse:
		gotoOp := ctx.AppendOp(vm.TXOPGoto, 0)
		pos := ctx.ByteCode.Len()
		for _, child := range n.(*node.ElseNode).ListNode.Nodes {
			c.compile(ctx, child)
		}
		gotoOp.SetArg(ctx.ByteCode.Len() - pos + 1)
	case node.NodeMakeArray:
		x := n.(*node.UnaryNode)
		c.compile(ctx, x.Child)
		ctx.AppendOp(vm.TXOPMakeArray)
	case node.NodeRange:
		x := n.(*node.BinaryNode)
		c.compile(ctx, x.Right)
		ctx.AppendOp(vm.TXOPPush)
		c.compile(ctx, x.Left)
		ctx.AppendOp(vm.TXOPMoveToSb)
		ctx.AppendOp(vm.TXOPPop)
		ctx.AppendOp(vm.TXOPRange)
	case node.NodeInt:
		x := n.(*node.NumberNode)
		ctx.AppendOp(vm.TXOPLiteral, x.Value.Int())
	case node.NodeList:
		x := n.(*node.ListNode)
		for _, v := range x.Nodes {
			c.compile(ctx, v)
			if v.Type() != node.NodeRange {
				ctx.AppendOp(vm.TXOPPush)
			}
		}
	case node.NodeFunCall:
		x := n.(*node.FunCallNode)

		for _, child := range x.Args.Nodes {
			c.compile(ctx, child)
			ctx.AppendOp(vm.TXOPPush)
		}

		c.compile(ctx, x.Invocant)
		ctx.AppendOp(vm.TXOPFunCallOmni)
	case node.NodeMethodCall:
		x := n.(*node.MethodCallNode)

		c.compile(ctx, x.Invocant)
		ctx.AppendOp(vm.TXOPPush)
		ctx.AppendOp(vm.TXOPPushmark)
		for _, child := range x.Args.Nodes {
			c.compile(ctx, child)
			ctx.AppendOp(vm.TXOPPush)
		}
		ctx.AppendOp(vm.TXOPMethodCall, x.MethodName)
		ctx.AppendOp(vm.TXOPPopmark)
	case node.NodeInclude:
		c.compileInclude(ctx, n.(*node.IncludeNode))
	case node.NodeGroup:
		c.compile(ctx, n.(*node.UnaryNode).Child)
	case node.NodeEquals, node.NodeNotEquals, node.NodeLT, node.NodeGT:
		x := n.(*node.BinaryNode)

		c.compileBinaryOperands(ctx, x)
		switch n.Type() {
		case node.NodeEquals:
			ctx.AppendOp(vm.TXOPEquals)
		case node.NodeNotEquals:
			ctx.AppendOp(vm.TXOPNotEquals)
		case node.NodeLT:
			ctx.AppendOp(vm.TXOPLessThan)
		case node.NodeGT:
			ctx.AppendOp(vm.TXOPGreaterThan)
		default:
			panic("Unknown operator")
		}
	case node.NodePlus, node.NodeMinus, node.NodeMul, node.NodeDiv:
		c.compileBinaryArithmetic(ctx, n.(*node.BinaryNode))
	case node.NodeFilter:
		x := n.(*node.FilterNode)

		c.compile(ctx, x.Child)
		ctx.AppendOp(vm.TXOPFilter, x.Name)
	case node.NodeWrapper:
		c.compileWrapper(ctx, n.(*node.WrapperNode))
	case node.NodeMacro:
		c.compileMacro(ctx, n.(*node.MacroNode))
	default:
		fmt.Printf("Unknown node: %s\n", n.Type())
	}
}

func (c *BasicCompiler) compileIf(ctx *context, n node.Node) {
	x := n.(*node.IfNode)
	ctx.AppendOp(vm.TXOPPushmark)
	c.compile(ctx, x.BooleanExpression)
	ifop := ctx.AppendOp(vm.TXOPAnd, 0)
	pos := ctx.ByteCode.Len()

	var elseNode node.Node
	children := x.ListNode.Nodes
	for _, child := range children {
		if child.Type() == node.NodeElse {
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

func (c *BasicCompiler) compileBinaryOperands(ctx *context, x *node.BinaryNode) {
	if x.Right.Type() == node.NodeGroup {
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

func (c *BasicCompiler) compileAssignmentNodes(ctx *context, assignnodes []node.Node) {
	if len(assignnodes) <= 0 {
		return
	}
	ctx.AppendOp(vm.TXOPPushmark)
	for _, nv := range assignnodes {
		v := nv.(*node.AssignmentNode)
		ctx.AppendOp(vm.TXOPLiteral, v.Assignee.Name)
		ctx.AppendOp(vm.TXOPPush)
		c.compile(ctx, v.Expression)
		ctx.AppendOp(vm.TXOPPush)
	}
	ctx.AppendOp(vm.TXOPMakeHash)
	ctx.AppendOp(vm.TXOPMoveToSb)
	ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileForeach(ctx *context, x *node.ForeachNode) {
	ctx.AppendOp(vm.TXOPPushmark)
	ctx.AppendOp(vm.TXOPPushFrame)
	c.compile(ctx, x.List)
	ctx.AppendOp(vm.TXOPForStart, x.IndexVarIdx)
	ctx.AppendOp(vm.TXOPLiteral, x.IndexVarIdx)

	iter := ctx.AppendOp(vm.TXOPForIter, 0)
	pos := ctx.ByteCode.Len()

	children := x.Nodes
	for _, v := range children {
		c.compile(ctx, v)
	}

	ctx.AppendOp(vm.TXOPGoto, -1*(ctx.ByteCode.Len()-pos+2))
	iter.SetArg(ctx.ByteCode.Len() - pos + 1)
	ctx.AppendOp(vm.TXOPPopFrame)
	ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileWhile(ctx *context, x *node.WhileNode) {
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
	ctx.AppendOp(vm.TXOPGoto, -1*(ctx.ByteCode.Len()-condPos+1))
	ifop.SetArg(ctx.ByteCode.Len() - ifPos + 1)
	ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileWrapper(ctx *context, x *node.WrapperNode) {
	// Save the current io.Writer to the stack
	// This also creates pushes a bytes.Buffer into the stack
	// so that following operations write to that buffer
	ctx.AppendOp(vm.TXOPSaveWriter)

	// From this place on, executed opcodes will write to a temporary
	// new output
	for _, v := range x.ListNode.Nodes {
		c.compile(ctx, v)
	}

	// Pop the original writer, and place it back to the output
	// Also push the output onto the stack
	ctx.AppendOp(vm.TXOPRestoreWriter)

	// Arguments to include (WITH foo = "bar") need to be evaulated
	// in the OUTER context, but the variables need to be set in the
	// include context
	c.compileAssignmentNodes(ctx, x.AssignmentNodes)

	// Popt the "content"
	ctx.AppendOp(vm.TXOPPop)
	ctx.AppendOp(vm.TXOPPushmark)
	ctx.AppendOp(vm.TXOPWrapper, x.WrapperName)
	ctx.AppendOp(vm.TXOPPopmark)
}

func (c *BasicCompiler) compileMacro(ctx *context, x *node.MacroNode) {
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
}

func (c *BasicCompiler) compileInclude(ctx *context, x *node.IncludeNode) {
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
}

func (c *BasicCompiler) compileBinaryArithmetic(ctx *context, x *node.BinaryNode) {
	c.compileBinaryOperands(ctx, x)
	switch x.Type() {
	case node.NodePlus:
		ctx.AppendOp(vm.TXOPAdd)
	case node.NodeMinus:
		ctx.AppendOp(vm.TXOPSub)
	case node.NodeMul:
		ctx.AppendOp(vm.TXOPMul)
	case node.NodeDiv:
		ctx.AppendOp(vm.TXOPDiv)
	default:
		panic("Unknown arithmetic")
	}
}
