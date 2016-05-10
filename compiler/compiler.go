package compiler

import (
	"fmt"
	"strconv"

	"github.com/lestrrat/go-xslate/node"
	"github.com/lestrrat/go-xslate/parser"
	"github.com/lestrrat/go-xslate/vm"
)

// AppendOp creates and appends a new op to the current set of ByteCode
func (ctx *context) AppendOp(o vm.OpType, args ...interface{}) vm.Op {
	return ctx.ByteCode.AppendOp(o, args...)
}

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
		compile(ctx, n)
	}

	// When we're done compiling, always append an END op
	ctx.ByteCode.AppendOp(vm.TXOPEnd)

	opt := &NaiveOptimizer{}
	opt.Optimize(ctx.ByteCode)

	ctx.ByteCode.Name = ast.Name
	return ctx.ByteCode, nil
}

func compile(ctx *context, n node.Node) {
	switch n.Type() {
	case node.Int, node.Text:
		compileLiteral(ctx, n)
	case node.FetchSymbol:
		ctx.AppendOp(vm.TXOPFetchSymbol, n.(*node.TextNode).Text)
	case node.FetchField:
		ffnode := n.(*node.FetchFieldNode)
		compile(ctx, ffnode.Container)
		ctx.AppendOp(vm.TXOPFetchFieldSymbol, ffnode.FieldName)
	case node.FetchArrayElement:
		compileFetchArrayElement(ctx, n.(*node.BinaryNode))
	case node.LocalVar:
		compileLoadLvar(ctx, n.(*node.LocalVarNode))
	case node.Assignment:
		compileAssignment(ctx, n.(*node.AssignmentNode))
	case node.Print:
		compile(ctx, n.(*node.ListNode).Nodes[0])
		ctx.AppendOp(vm.TXOPPrint)
	case node.PrintRaw:
		compile(ctx, n.(*node.ListNode).Nodes[0])
		ctx.AppendOp(vm.TXOPPrintRaw)
	case node.Foreach:
		compileForeach(ctx, n.(*node.ForeachNode))
	case node.While:
		compileWhile(ctx, n.(*node.WhileNode))
	case node.If:
		compileIf(ctx, n)
	case node.Else:
		compileElse(ctx, n.(*node.ElseNode))
	case node.MakeArray:
		x := n.(*node.UnaryNode)
		compile(ctx, x.Child)
		ctx.AppendOp(vm.TXOPMakeArray)
	case node.Range:
		x := n.(*node.BinaryNode)
		compile(ctx, x.Right)
		ctx.AppendOp(vm.TXOPPush)
		compile(ctx, x.Left)
		ctx.AppendOp(vm.TXOPMoveToSb)
		ctx.AppendOp(vm.TXOPPop)
		ctx.AppendOp(vm.TXOPRange)
	case node.List:
		compileList(ctx, n.(*node.ListNode))
	case node.FunCall:
		compileFunCall(ctx, n.(*node.FunCallNode))
	case node.MethodCall:
		compileMethodCall(ctx, n.(*node.MethodCallNode))
	case node.Include:
		compileInclude(ctx, n.(*node.IncludeNode))
	case node.Group:
		compile(ctx, n.(*node.UnaryNode).Child)
	case node.Equals, node.NotEquals, node.LT, node.GT:
		compileComparison(ctx, n.(*node.BinaryNode))
	case node.Plus, node.Minus, node.Mul, node.Div:
		compileBinaryArithmetic(ctx, n.(*node.BinaryNode))
	case node.Filter:
		x := n.(*node.FilterNode)

		compile(ctx, x.Child)
		ctx.AppendOp(vm.TXOPFilter, x.Name)
	case node.Wrapper:
		compileWrapper(ctx, n.(*node.WrapperNode))
	case node.Macro:
		compileMacro(ctx, n.(*node.MacroNode))
	default:
		fmt.Printf("Unknown node: %s\n", n.Type())
	}
}

func compileComparison(ctx *context, n *node.BinaryNode) {
	compileBinaryOperands(ctx, n)
	switch n.Type() {
	case node.Equals:
		ctx.AppendOp(vm.TXOPEquals)
	case node.NotEquals:
		ctx.AppendOp(vm.TXOPNotEquals)
	case node.LT:
		ctx.AppendOp(vm.TXOPLessThan)
	case node.GT:
		ctx.AppendOp(vm.TXOPGreaterThan)
	default:
		panic("Unknown operator")
	}
}

func compileFetchArrayElement(ctx *context, n *node.BinaryNode) {
	ctx.AppendOp(vm.TXOPPushmark).SetComment("fetch array element")
	compile(ctx, n.Right)
	ctx.AppendOp(vm.TXOPPush)
	compile(ctx, n.Left)
	ctx.AppendOp(vm.TXOPPush)
	ctx.AppendOp(vm.TXOPFetchArrayElement)
	ctx.AppendOp(vm.TXOPPopmark)
}

func compileFunCall(ctx *context, n *node.FunCallNode) {
	if len(n.Args.Nodes) > 0 {
		ctx.AppendOp(vm.TXOPNoop).SetComment("Setting up function arguments")
		for _, child := range n.Args.Nodes {
			compile(ctx, child)
			ctx.AppendOp(vm.TXOPPush)
		}
	}

	if inv := n.Invocant; inv != nil {
		compile(ctx, inv)
	}
	ctx.AppendOp(vm.TXOPFunCallOmni)
}

func compileMethodCall(ctx *context, n *node.MethodCallNode) {
	ctx.AppendOp(vm.TXOPPushmark).SetComment("Begin method call")
	compile(ctx, n.Invocant)
	ctx.AppendOp(vm.TXOPPush).SetComment("Push method invocant")
	for _, child := range n.Args.Nodes {
		compile(ctx, child)
		ctx.AppendOp(vm.TXOPPush)
	}
	ctx.AppendOp(vm.TXOPMethodCall, n.MethodName)
	ctx.AppendOp(vm.TXOPPopmark).SetComment("End method call")
}

func compileList(ctx *context, n *node.ListNode) {
	ctx.AppendOp(vm.TXOPNoop).SetComment("BEGIN list")
	for _, v := range n.Nodes {
		compile(ctx, v)
		if v.Type() != node.Range {
			ctx.AppendOp(vm.TXOPPush)
		}
	}
	ctx.AppendOp(vm.TXOPNoop).SetComment("END list")
}

func compileIf(ctx *context, n node.Node) {
	x := n.(*node.IfNode)
	ctx.AppendOp(vm.TXOPPushmark).SetComment("BEGIN IF")
	compile(ctx, x.BooleanExpression)
	ifop := ctx.AppendOp(vm.TXOPAnd, 0)
	pos := ctx.ByteCode.Len()

	var elseNode node.Node
	children := x.ListNode.Nodes
	for _, child := range children {
		if child.Type() == node.Else {
			elseNode = child
		} else {
			compile(ctx, child)
		}
	}

	if elseNode == nil {
		ifop.SetArg(ctx.ByteCode.Len() - pos + 1)
		ifop.SetComment("Jump to end of IF at " + strconv.Itoa(ctx.ByteCode.Len()+1) + " when condition fails")
	} else {
		// If we have an else, we need to put this AFTER the goto
		// that's generated by else
		ifop.SetArg(ctx.ByteCode.Len() - pos + 2)
		ifop.SetComment("Jump to ELSE at " + strconv.Itoa(ctx.ByteCode.Len()+2) + " when condition fails")
		compile(ctx, elseNode)
	}
	ctx.AppendOp(vm.TXOPPopmark).SetComment("END IF")

}

func compileElse(ctx *context, n *node.ElseNode) {
	gotoOp := ctx.AppendOp(vm.TXOPGoto, 0)
	pos := ctx.ByteCode.Len()
	for _, child := range n.ListNode.Nodes {
		compile(ctx, child)
	}
	gotoOp.SetArg(ctx.ByteCode.Len() - pos + 1)
}

func compileBinaryOperands(ctx *context, x *node.BinaryNode) {
	if x.Right.Type() == node.Group {
		// Grouped node
		compile(ctx, x.Right)
		ctx.AppendOp(vm.TXOPPush)
		compile(ctx, x.Left)
		ctx.AppendOp(vm.TXOPMoveToSb)
		ctx.AppendOp(vm.TXOPPop)
	} else {
		compile(ctx, x.Left)
		ctx.AppendOp(vm.TXOPMoveToSb)
		compile(ctx, x.Right)
	}
}

func compileAssignmentNodes(ctx *context, assignnodes []node.Node) {
	if len(assignnodes) <= 0 {
		return
	}
	ctx.AppendOp(vm.TXOPPushmark)
	for _, nv := range assignnodes {
		v := nv.(*node.AssignmentNode)
		ctx.AppendOp(vm.TXOPLiteral, v.Assignee.Name)
		ctx.AppendOp(vm.TXOPPush)
		compile(ctx, v.Expression)
		ctx.AppendOp(vm.TXOPPush)
	}
	ctx.AppendOp(vm.TXOPMakeHash)
	ctx.AppendOp(vm.TXOPMoveToSb)
	ctx.AppendOp(vm.TXOPPopmark)
}

func compileForeach(ctx *context, x *node.ForeachNode) {
	ctx.AppendOp(vm.TXOPPushmark).SetComment("BEGIN FOREACH")
	ctx.AppendOp(vm.TXOPPushFrame).SetComment("BEGIN new scope")
	compile(ctx, x.List)
	ctx.AppendOp(vm.TXOPForStart, x.IndexVarIdx)
	ctx.AppendOp(vm.TXOPLiteral, x.IndexVarIdx)

	iter := ctx.AppendOp(vm.TXOPForIter, 0)
	pos := ctx.ByteCode.Len()

	children := x.Nodes
	for _, v := range children {
		compile(ctx, v)
	}

	ctx.AppendOp(vm.TXOPGoto, -1*(ctx.ByteCode.Len()-pos+2)).SetComment("Jump back to for_iter at " + strconv.Itoa(pos))
	ctx.AppendOp(vm.TXOPPopFrame).SetComment("END scope")

	// Tell for iter to jump to this position when
	// the loop is done.
	iter.SetArg(ctx.ByteCode.Len()-pos)
	iter.SetComment("Jump to end of scope at " + strconv.Itoa(ctx.ByteCode.Len()) + " when we're done")
	ctx.AppendOp(vm.TXOPPopmark).SetComment("END FOREACH")
}

func compileWhile(ctx *context, x *node.WhileNode) {
	ctx.AppendOp(vm.TXOPPushmark)
	ctx.AppendOp(vm.TXOPPushFrame)
	ctx.AppendOp(vm.TXOPLiteral, 0)
	ctx.AppendOp(vm.TXOPSaveToLvar, 0)

	condPos := ctx.ByteCode.Len() + 1

	// compile the boolean expression
	compile(ctx, x.Condition)

	// we might as well use the equivalent of If here!
	ifop := ctx.AppendOp(vm.TXOPAnd, 0)
	ifPos := ctx.ByteCode.Len()

	children := x.Nodes
	for _, v := range children {
		compile(ctx, v)
	}

	// Go back to condPos
	ctx.AppendOp(vm.TXOPGoto, -1*(ctx.ByteCode.Len()-condPos+1)).SetComment("Jump to " + strconv.Itoa(condPos))
	ifop.SetArg(ctx.ByteCode.Len() - ifPos + 1)
	ifop.SetComment("Jump to " + strconv.Itoa(ctx.ByteCode.Len() + 1))
	ctx.AppendOp(vm.TXOPPopmark)
}

func compileWrapper(ctx *context, x *node.WrapperNode) {
	// Save the current io.Writer to the stack
	// This also creates pushes a bytes.Buffer into the stack
	// so that following operations write to that buffer
	ctx.AppendOp(vm.TXOPSaveWriter)

	// From this place on, executed opcodes will write to a temporary
	// new output
	for _, v := range x.ListNode.Nodes {
		compile(ctx, v)
	}

	// Pop the original writer, and place it back to the output
	// Also push the output onto the stack
	ctx.AppendOp(vm.TXOPRestoreWriter)

	// Arguments to include (WITH foo = "bar") need to be evaulated
	// in the OUTER context, but the variables need to be set in the
	// include context
	compileAssignmentNodes(ctx, x.AssignmentNodes)

	// Popt the "content"
	ctx.AppendOp(vm.TXOPPop)
	ctx.AppendOp(vm.TXOPPushmark)
	ctx.AppendOp(vm.TXOPWrapper, x.WrapperName)
	ctx.AppendOp(vm.TXOPPopmark)
}

func compileMacro(ctx *context, x *node.MacroNode) {
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
		compile(ctx, child)
	}
	ctx.AppendOp(vm.TXOPPopmark)
	ctx.AppendOp(vm.TXOPEnd) // This END forces termination
	gotoOp.SetArg(ctx.ByteCode.Len() - start + 1)

	// Now remember about this definition
	ctx.AppendOp(vm.TXOPLiteral, entryPoint)
	ctx.AppendOp(vm.TXOPSaveToLvar, x.LocalVar.Offset)
}

func compileInclude(ctx *context, x *node.IncludeNode) {
	compile(ctx, x.IncludeTarget)
	ctx.AppendOp(vm.TXOPPush)
	// Arguments to include (WITH foo = "bar") need to be evaulated
	// in the OUTER context, but the variables need to be set in the
	// include context
	compileAssignmentNodes(ctx, x.AssignmentNodes)
	ctx.AppendOp(vm.TXOPPop)
	ctx.AppendOp(vm.TXOPPushmark)
	ctx.AppendOp(vm.TXOPInclude)
	ctx.AppendOp(vm.TXOPPopmark)
}

func compileBinaryArithmetic(ctx *context, n *node.BinaryNode) {
	var optype vm.OpType
	switch n.Type() {
	case node.Plus:
		optype = vm.TXOPAdd
	case node.Minus:
		optype = vm.TXOPSub
	case node.Mul:
		optype = vm.TXOPMul
	case node.Div:
		optype = vm.TXOPDiv
	default:
		panic("Unknown arithmetic")
	}
	ctx.AppendOp(vm.TXOPNoop).SetComment("BEGIN " + optype.String())
	compileBinaryOperands(ctx, n)
	ctx.AppendOp(optype).SetComment("Execute " + optype.String() + " on registers sa and sb")
}

func compileLiteral(ctx *context, n node.Node) {
	var op vm.Op
	switch n.Type() {
	case node.Int:
		op = ctx.AppendOp(vm.TXOPLiteral, n.(*node.NumberNode).Value.Int())
	case node.Text:
		op = ctx.AppendOp(vm.TXOPLiteral, n.(*node.TextNode).Text)
	default:
		panic("unknown literal value")
	}
	op.SetComment("Save literal to sa")
}

func compileAssignment(ctx *context, n *node.AssignmentNode) {
	compile(ctx, n.Expression)
	// XXX this 0 must be pre-computed
	ctx.AppendOp(vm.TXOPSaveToLvar, 0).SetComment("Saving to local var 0")
}

func compileLoadLvar(ctx *context, n *node.LocalVarNode) {
	ctx.AppendOp(vm.TXOPLoadLvar, n).SetComment("Load variable '" + n.Name + "' to sa")
}
