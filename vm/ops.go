package vm

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/lestrrat-go/xslate/functions"
	"github.com/lestrrat-go/xslate/functions/array"
	"github.com/lestrrat-go/xslate/functions/hash"
	"github.com/lestrrat-go/xslate/internal/rbpool"
	"github.com/lestrrat-go/xslate/internal/rvpool"
	"github.com/lestrrat-go/xslate/node"
)

func init() {
	for i := TXOPNoop; i < TXOPMax; i++ {
		var h OpHandler
		n := "Unknown"
		switch i {
		case TXOPNoop:
			h = txNoop
			n = "noop"
		case TXOPEnd:
			h = txEnd
			n = "end"
		case TXOPMoveToSb:
			h = txMoveToSb
			n = "move_to_sb"
		case TXOPMoveFromSb:
			h = txMoveFromSb
			n = "move_from_sb"
		case TXOPMarkRaw:
			h = txMarkRaw
			n = "mark_raw"
		case TXOPUnmarkRaw:
			h = txUnmarkRaw
			n = "unmark_raw"
		case TXOPPrint:
			h = txPrint
			n = "print"
		case TXOPPrintRaw:
			h = txPrintRaw
			n = "print_raw"
		case TXOPPrintRawConst:
			h = txPrintRawConst
			n = "print_raw_const"
		case TXOPLiteral:
			h = txLiteral
			n = "literal"
		case TXOPFetchSymbol:
			h = txFetchSymbol
			n = "fetch_s"
		case TXOPFetchFieldSymbol:
			h = txFetchField
			n = "fetch_field_s"
		case TXOPFetchArrayElement:
			h = txFetchArrayElement
			n = "fetch_array_elem"
		case TXOPSaveToLvar:
			h = txSaveToLvar
			n = "save_to_lvar"
		case TXOPLoadLvar:
			h = txLoadLvar
			n = "load_lvar"
		case TXOPNil:
			h = txNil
			n = "nil"
		case TXOPAdd:
			h = txAdd
			n = "add"
		case TXOPSub:
			h = txSub
			n = "sub"
		case TXOPMul:
			h = txMul
			n = "mul"
		case TXOPDiv:
			h = txDiv
			n = "div"
		case TXOPAnd:
			h = txAnd
			n = "and"
		case TXOPGoto:
			h = txGoto
			n = "goto"
		case TXOPForStart:
			h = txForStart
			n = "for_start"
		case TXOPForIter:
			h = txForIter
			n = "for_iter"
		case TXOPHTMLEscape:
			h = txHTMLEscape
			n = "html_escape"
		case TXOPUriEscape:
			h = txUriEscape
			n = "uri_escape"
		case TXOPEquals:
			h = txEquals
			n = "equals"
		case TXOPNotEquals:
			h = txNotEquals
			n = "not_equals"
		case TXOPLessThan:
			h = txLessThan
			n = "less_than"
		case TXOPGreaterThan:
			h = txGreaterThan
			n = "greater_than"
		case TXOPPush:
			h = txPush
			n = "push"
		case TXOPPop:
			h = txPop
			n = "pop"
		case TXOPPopmark:
			h = txPopmark
			n = "popmark"
		case TXOPPushmark:
			h = txPushmark
			n = "pushmark"
		case TXOPPopFrame:
			h = txPopFrame
			n = "popframe"
		case TXOPPushFrame:
			h = txPushFrame
			n = "pushframe"
		case TXOPFunCall:
			h = txFunCall
			n = "funcall"
		case TXOPFunCallSymbol:
			h = txFunCallSymbol
			n = "funcall_symbol"
		case TXOPFunCallOmni:
			h = txFunCallOmni
			n = "funcall_omni"
		case TXOPMethodCall:
			h = txMethodCall
			n = "methodcall"
		case TXOPRange:
			h = txRange
			n = "range"
		case TXOPMakeArray:
			h = txMakeArray
			n = "make_array"
		case TXOPMakeHash:
			h = txMakeHash
			n = "make_hash"
		case TXOPInclude:
			h = txInclude
			n = "include"
		case TXOPWrapper:
			h = txWrapper
			n = "wrapper"
		case TXOPFilter:
			h = txFilter
			n = "filter"
		case TXOPSaveWriter:
			h = txSaveWriter
			n = "save_writer"
		case TXOPRestoreWriter:
			h = txRestoreWriter
			n = "restore_writer"
		default:
			panic("No such optype")
		}
		ophandlers[i] = h
		opnames[i] = n
	}
}

func optypeToHandler(o OpType) OpHandler {
	return ophandlers[o]
}

func txEnd(st *State) {}

func txNil(st *State) {
	st.sa = nil
	st.Advance()
}

func txNoop(st *State) {
	st.Advance()
}

// Moves content of register sa to register sb
func txMoveToSb(st *State) {
	st.sb = st.sa
	st.Advance()
}

// Moves content of register sb to register sa
func txMoveFromSb(st *State) {
	st.sa = st.sb
	st.Advance()
}

// Sets literal in op arg to register sa
func txLiteral(st *State) {
	st.sa = st.CurrentOp().Arg()
	st.Advance()
}

// Fetches a symbol specified in op arg from template variables.
// XXX need to handle local vars?
func txFetchSymbol(st *State) {
	// Need to handle local vars?
	key := st.CurrentOp().Arg()
	vars := st.Vars()
	if v, ok := vars.Get(key); ok {
		st.sa = v
	} else {
		st.sa = nil
	}
	st.Advance()
}

// pushmark
// load_lvar 0
// push
// literal_i start
// push
// literal_i end
// push
// fetch_slice
/*
func txFetchSlice(st *State) {
  container := st.sa
  if container == nil {
    // XXX ? no op?
    st.sa = nil
  } else {
    v := reflect.ValueOf(container)
    v.Slice(
*/

func txFetchField(st *State) {
	container := st.sa
	if container == nil {
		// XXX ? no op?
		st.sa = nil
	} else {
		t := reflect.TypeOf(container)
		var f reflect.Value
		var v reflect.Value
		name := st.CurrentOp().ArgString()
		switch t.Kind() {
		case reflect.Ptr, reflect.Struct:
			// Uppercase first character of field name
			r, size := utf8.DecodeRuneInString(name)
			name = string(unicode.ToUpper(r)) + name[size:]

			v = reflect.ValueOf(container)
			if t.Kind() == reflect.Ptr {
				// dereference
				v = v.Elem()
			}

			if v.Type().Name() == "LoopVar" {
				// some special treatment here
				switch name {
				case "Max":
					name = "MaxIndex"
				case "Next":
					name = "PeekNext"
				case "Prev":
					name = "PeeekPrev"
				case "First":
					name = "IsFirst"
				case "Last":
					name = "IsLast"
				}
			}

			f = v.FieldByName(name)
		case reflect.Map:
			v = reflect.ValueOf(container)
			f = v.MapIndex(reflect.ValueOf(name))
		default:
			panic(fmt.Sprintf("XXX Put proper error handling here: %s (%s)", container, t))
		}

		st.sa = f.Interface()
	}
	st.Advance()
}

func txFetchArrayElement(st *State) {
	defer st.Advance()

	array := reflect.ValueOf(st.StackPop())
	switch array.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		st.Warnf("cannot index into non-array/slice element")
		return
	}

	idx := st.StackPop()
	v := array.Index(int(idx.(int64)))
	st.sa = v.Interface()
}

type rawString string

func (s rawString) String() string { return string(s) }

var rawStringType = reflect.TypeOf(new(rawString)).Elem()

// Wraps the contents of register sa with a "raw string" mark
// Note that this effectively stringifies the contents of register sa
func txMarkRaw(st *State) {
	if reflect.ValueOf(st.sa).Type() != rawStringType {
		st.sa = rawString(interfaceToString(st.sa))
	}
	st.Advance()
}

// Sets the contents of register sa to a regular string, and removes
// the "raw string" mark, forcing html escapes to be applied when printing.
// Note that this effectively stringifies the contents of register sa
func txUnmarkRaw(st *State) {
	if reflect.ValueOf(st.sa).Type() == rawStringType {
		st.sa = string(interfaceToString(st.sa))
	}
	st.Advance()
}

// Prints the contents of register sa to Output.
// Forcefully applies html escaping unless the variable in sa is marked "raw"
func txPrint(st *State) {
	arg := st.sa
	if arg == nil {
		st.Warnf("Use of nil to print\n")
	} else if reflect.ValueOf(st.sa).Type() != rawStringType {
		st.AppendOutputString(html.EscapeString(interfaceToString(arg)))
	} else {
		st.AppendOutputString(interfaceToString(arg))
	}
	st.Advance()
}

func txPrintRawConst(st *State) {
	st.AppendOutputString(st.CurrentOp().ArgString())
	st.Advance()
}

// Prints the contents of register sa, forcing raw string semantics
func txPrintRaw(st *State) {
	// XXX TODO: mark_raw handling
	arg := st.sa
	if arg == nil {
		st.Warnf("Use of nil to print\n")
	} else {
		st.AppendOutputString(interfaceToString(arg))
	}
	st.Advance()
}

func txSaveToLvar(st *State) {
	idx := st.CurrentOp().ArgInt()
	st.CurrentFrame().SetLvar(idx, st.sa)
	st.Advance()
}

func txLoadLvar(st *State) {
	n := st.CurrentOp().Arg().(*node.LocalVarNode)
	v, err := st.CurrentFrame().GetLvar(n.Offset)
	if err != nil {
		st.Warnf("failed to load variable '%s': %s\n", n.Name, err)
	} else {
		st.sa = v
	}
	st.Advance()
}

func txAdd(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.sa = leftV.Int() + rightV.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() + rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() + rightV.Float()
	}

	// XXX: set to targ?
	st.Advance()
}

func txSub(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.sa = leftV.Int() - rightV.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() - rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() - rightV.Float()
	}

	// XXX: set to targ?
	st.Advance()
}

func txMul(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.sa = leftV.Int() * rightV.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() * rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() * rightV.Float()
	}

	// XXX: set to targ?
	st.Advance()
}

func txDiv(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// XXX This is a hack. We rely on interfaceToString() using FormatFloat(prec = -1)
		// to get rid of the fractional portions when printing
		typeF := reflect.TypeOf(0.1)
		st.sa = leftV.Convert(typeF).Float() / rightV.Convert(typeF).Float()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() / rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() / rightV.Float()
	}

	// XXX: set to targ?
	st.Advance()
}

func txAnd(st *State) {
	if interfaceToBool(st.sa) {
		st.Advance()
	} else {
		st.AdvanceBy(st.CurrentOp().ArgInt())
	}
}

func txGoto(st *State) {
	st.AdvanceBy(st.CurrentOp().ArgInt())
}

// NewLoopVar creates the loop variable
func NewLoopVar(idx int, array reflect.Value) *LoopVar {
	lv := &LoopVar{
		Index:    idx,
		Count:    idx + 1,
		Body:     array,
		Size:     array.Len(),
		MaxIndex: array.Len() - 1,
		PeekNext: nil,
		PeekPrev: nil,
		IsFirst:  false,
		IsLast:   false,
	}
	return lv
}

func txForStart(st *State) {
	array := reflect.ValueOf(st.sa)

	switch array.Kind() {
	case reflect.Array, reflect.Slice:
		// Normal case. nothing to do
	default:
		// Oh you silly goose. You didn't give me a array.
		// Use a dummy array
		array = reflect.ValueOf([]struct{}{})
	}

	cf := st.CurrentFrame()
	cf.SetLvar(0, nil) // item
	cf.SetLvar(1, NewLoopVar(-1, array))

	st.Advance()
}

func txForIter(st *State) {
	cf := st.CurrentFrame()
	var loop *LoopVar

	// The loop variable MUST exist. Not having one is a sure panic
	v, err := cf.GetLvar(1)
	if err != nil {
		panic("loop var not found: " + err.Error())
	}

	var ok bool
	if loop, ok = v.(*LoopVar); !ok {
		panic("failed to convert loop var")
	}

	slice := loop.Body
	loop.Index++
	loop.Count++
	if loop.Count > st.MaxLoopCount {
		panic("looped for " + strconv.Itoa(loop.Count) + " times, aborting")
	}

	loop.IsFirst = loop.Index == 0
	loop.IsLast = loop.Index == loop.MaxIndex

	if loop.Size > loop.Index {
		cf.SetLvar(0, slice.Index(loop.Index).Interface())

		if loop.Size > loop.Index+1 {
			loop.PeekNext = slice.Index(loop.Index + 1).Interface()
		} else {
			loop.PeekNext = nil
		}

		if loop.Index > 0 {
			loop.PeekPrev = slice.Index(loop.Index - 1).Interface()
		} else {
			loop.PeekPrev = nil
		}
		st.Advance()
		return
	}

	// loop done
	st.AdvanceBy(st.CurrentOp().ArgInt())
}

func txFilter(st *State) {
	name := st.CurrentOp().Arg().(string)

	// XXX Check for local vars first?
	switch name {
	case "html":
		txHTMLEscape(st)
	case "uri":
		txUriEscape(st)
	case "mark_raw":
		txMarkRaw(st)
	default:
		panic("User-specified filters not implemented yet")
	}
}

func txUriEscape(st *State) {
	v := interfaceToString(st.sa)
	st.sa = escapeUriString(v)
	st.Advance()
}

func txHTMLEscape(st *State) {
	v := interfaceToString(st.sa)
	st.sa = rawString(html.EscapeString(v))
	st.Advance()
}

func _txEquals(st *State) bool {
	var leftV, rightV interface{}

	switch {
	case isInterfaceNumeric(st.sb):
		leftV, rightV = alignTypesForArithmetic(st.sb, st.sa)
	case isInterfaceStringType(st.sb):
		leftV, rightV = interfaceToString(st.sb), interfaceToString(st.sa)
	default:
		leftV, rightV = st.sb, st.sa
	}

	switch leftV.(type) {
	case reflect.Value:
		leftVV := leftV.(reflect.Value)
		rightVV := rightV.(reflect.Value)
		switch leftVV.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return leftVV.Int() == rightVV.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return leftVV.Uint() == rightVV.Uint()
		case reflect.Float32, reflect.Float64:
			return leftVV.Float() == rightVV.Float()
		default:
			panic(fmt.Sprintf("Unhandled type in '==': %s", leftVV.Kind()))
		}
	default:
		return leftV == rightV
	}
}

func txEquals(st *State) {
	st.sa = _txEquals(st)
	st.Advance()
}

func txNotEquals(st *State) {
	st.sa = !_txEquals(st)
	st.Advance()
}

func txLessThan(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.sa = leftV.Int() < rightV.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() < rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() < rightV.Float()
	}
	st.Advance()
}

func txGreaterThan(st *State) {
	leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
	switch leftV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.sa = leftV.Int() > rightV.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.sa = leftV.Uint() > rightV.Uint()
	case reflect.Float32, reflect.Float64:
		st.sa = leftV.Float() > rightV.Float()
	}
	st.Advance()
}

// func/method call related stuff
// Note: You MUST MUST MUST call pushmark before setting up the argument
// list on the stack
/*
In the original p5-Text-Xslate, foo.hoge(1, 2, 3) generates the
following bytecode:

  pushmark // hoge
  load_lvar 0 #2
  push
  literal_i 1
  push
  literal_i 2
  push
  literal_i 3
  push
  methodcall_s "hoge" #2

*/
func txPopmark(st *State) {
	st.Popmark()
	st.Advance()
}

func txPushmark(st *State) {
	st.Pushmark()
	st.Advance()
}

func txPushFrame(st *State) {
	st.PushFrame()
	st.Advance()
}

func txPopFrame(st *State) {
	st.PopFrame()
	st.Advance()
}

func txPush(st *State) {
	st.StackPush(st.sa)
	st.Advance()
}

func txPop(st *State) {
	st.sa = st.StackPop()
	st.Advance()
}

var funcZero = reflect.Zero(reflect.ValueOf(func() {}).Type())

func invokeFuncSingleReturn(st *State, fun reflect.Value, args []reflect.Value) {
	if fun.Type().NumIn() != len(args) {
		st.Warnf("Number of arguments for function does not match (expected %d, got %d)\n", fun.Type().NumIn(), len(args))
		st.sa = ""
	} else if fun.Type().NumOut() == 0 {
		// Purely for side effect
		st.sa = ""
	} else {
		ret := fun.Call(args)
		// grab only the first return value. If you need the
		// entire return value set, you need to call invokeFunMultiReturn
		// (to be implemented)
		st.sa = ret[0].Interface()
	}
}

// Function calls (NOT to be confused with method calls, which are totally
// sane, and fine) in go-xslate is a bit different. Unlike in non-compiled
// languages like Perl, we can't just lookup a function out of nowhere
// with just its name: Golang's reflection mechanism requires that you
// have a concret value before doing a lookup.
//
// So this is not possible (where "time" is just a string):
//
//  [% time.Now() %]
//
// However, it's possible to register the function itself as a variable:
//
//  tx.Render(..., map[string]interface {} { "now": time.Now })
//  [% now() %]
//
// But this requires function names to be globally unique. That's not always
// possible. To avoid this, we can register an OBJECT named "time" before hand
// so that it in turn calls the ordinary time.Now() function
//
//  // Exact usage TBD
//  tx.RegisterFunctions(txtime.New())
//  tx.Render(...)
//  [% time.Now %]
// ...And that's how we manage function calls
// See also:
func txFunCall(st *State) {
	// Everything in our lvars up to the current tip is our argument list
	mark := st.CurrentMark()
	tip := st.stack.Size() - 1
	var args []reflect.Value

	if tip-mark-1 > 0 {
		args = make([]reflect.Value, tip-mark-1)
		for i := mark + 1; tip > i; i++ {
			v, _ := st.stack.Get(i)
			args[i-mark] = reflect.ValueOf(v)
		}
	}

	x := st.sa
	st.sa = nil
	if x == nil {
		// Do nothing, just advance
		st.Advance()
		return
	}

	v := reflect.ValueOf(x)
	if v.Type().Kind() == reflect.Func {
		fun := reflect.ValueOf(x)
		invokeFuncSingleReturn(st, fun, args)
	}
	st.Advance()
}

func txFunCallSymbol(st *State) {
	// Everything in our lvars up to the current tip is our argument list
	mark := st.CurrentMark()
	tip := st.stack.Size() - 1
	var args []reflect.Value

	if tip-mark-1 > 0 {
		args = make([]reflect.Value, tip-mark-1)
		for i := mark + 1; tip > i; i++ {
			v, _ := st.stack.Get(i)
			args[i-mark] = reflect.ValueOf(v)
		}
	}

	x, _ := st.stack.Get(mark)
	v := reflect.ValueOf(x)

	st.sa = nil
	if st.CurrentOp().Arg() != nil {
		vtype := v.Type()
		if vtype.Kind() == reflect.Ptr && vtype.Elem().Kind() == reflect.Struct && v.Elem().Type().Name() == "FuncDepot" {
			name := interfaceToString(st.CurrentOp().Arg())
			fd := x.(*functions.FuncDepot)
			fun, ok := fd.Get(name)
			if ok {
				invokeFuncSingleReturn(st, fun, args)
			}
		}
	}
	st.Advance()
}

func txMethodCall(st *State) {
	defer st.Advance() // We advance, regardless of errors

	name := interfaceToString(st.CurrentOp().Arg())
	// Uppercase first character of field name
	r, size := utf8.DecodeRuneInString(name)
	name = string(unicode.ToUpper(r)) + name[size:]

	mark := st.CurrentMark()
	tip := st.stack.Size()

	var invocant reflect.Value
	args := make([]reflect.Value, tip-mark)
	for i := mark; i < tip; i++ {
		v := st.stack.Pop()
		args[tip-i-1] = reflect.ValueOf(v)
		if i == mark {
			invocant = args[tip-i-1]
		}
	}

	// For maps, arrays, slices, we call virtual methods, if they are available
	switch invocant.Kind() {
	case reflect.Map:
		fun, ok := hash.Depot().Get(name)
		if ok {
			invokeFuncSingleReturn(st, fun, args)
		}
	case reflect.Array, reflect.Slice:
		// Array/Slices cannot be passed as []interface {} or any other
		// generic way, so we need to re-allocate the argument to a more
		// generic []interface {} container before dispatching it
		list := make([]interface{}, invocant.Len())
		for i := 0; i < invocant.Len(); i++ {
			list[i] = invocant.Index(i).Interface()
		}
		args[0] = reflect.ValueOf(list)
		fun, ok := array.Depot().Get(name)
		if ok {
			invokeFuncSingleReturn(st, fun, args)
		}
	default:
		method, ok := invocant.Type().MethodByName(name)
		if !ok {
			st.sa = nil
		} else {
			invokeFuncSingleReturn(st, method.Func, args)
		}
	}
}

// XXX can I just push a []int to st.sa?
func txRange(st *State) {
	lhs := interfaceToNumeric(st.sb).Int()
	rhs := interfaceToNumeric(st.sa).Int()

	for i := lhs; i <= rhs; i++ {
		// push these to stack
		st.StackPush(i)
	}

	st.Advance()
}

// Grab every thing from current mark up to the tip of the stack,
// and make it a list
func txMakeArray(st *State) {
	start := st.CurrentMark() // start
	end := st.StackTip()      // end

	if end <= start {
		panic(fmt.Sprintf("MakeArray: list start (%d) >= end (%d)", start, end))
	}

	list := make([]interface{}, end-start+1)
	for i := end; i >= start; i-- {
		list[i-start] = st.StackPop()
	}
	st.sa = list
	st.Advance()
}

func txMakeHash(st *State) {
	start := st.CurrentMark() // start
	end := st.StackTip()      // end

	hash := make(map[interface{}]interface{})
	for i := end; i > start; {
		v := st.StackPop()
		k := st.StackPop()
		hash[k] = v
		i -= 2
	}

	st.sa = hash
	st.Advance()
}

func txInclude(st *State) {
	// st.sa should contain the include target
	// st.sb should contain the map[interface{}]interface{}
	//   object that gets passed to the included template

	vars := Vars(rvpool.Get())
	defer rvpool.Release(vars)
	defer vars.Reset()
	if x := st.Vars(); x != nil {
		for k, v := range x {
			vars.Set(k, v)
		}
	}

	if x := st.sb; x != nil {
		hash := x.(map[interface{}]interface{})
		// Need to covert this to Vars (map[string]interface{})
		for k, v := range hash {
			vars.Set(interfaceToString(k), v)
		}
	}

	target := interfaceToString(st.sa)
	bc, err := st.LoadByteCode(target)
	if err != nil {
		panic(fmt.Sprintf("Include: Failed to compile %s: %s", target, err))
	}

	buf := rbpool.Get()
	defer rbpool.Release(buf)

	vm := NewVM()
	vm.Run(bc, vars, buf)
	st.AppendOutputString(buf.String())
	st.Advance()
}

func txWrapper(st *State) {
	// See txInclude
	vars := Vars(rvpool.Get())
	defer rvpool.Release(vars)
	defer vars.Reset()

	if x := st.Vars(); x != nil {
		for k, v := range x {
			vars.Set(k, v)
		}
	}

	if x := st.sb; x != nil {
		hash := x.(map[interface{}]interface{})
		// Need to covert this to Vars (map[string]interface{})
		for k, v := range hash {
			vars.Set(interfaceToString(k), v)
		}
	}
	vars.Set("content", rawString(st.sa.(string)))

	target := st.CurrentOp().ArgString()
	bc, err := st.LoadByteCode(target)
	if err != nil {
		panic(fmt.Sprintf("Wrapper: Failed to compile %s: %s", target, err))
	}

	vm := NewVM()
	vm.Run(bc, vars, st.output)
	st.Advance()
}

func txSaveWriter(st *State) {
	st.StackPush(st.output)

	buf := rbpool.Get()

	st.StackPush(buf)
	st.output = bufio.NewWriter(buf)
	st.Advance()
}

func txRestoreWriter(st *State) {
	st.output.(*bufio.Writer).Flush()
	buf := st.StackPop().(*bytes.Buffer)
	st.output = st.StackPop().(io.Writer)

	st.StackPush(buf.String())
	rbpool.Release(buf)

	st.Advance()
}

func txMacroCall(st *State) {
	x := st.sa.(int)
	bc := NewByteCode()
	bc.OpList = st.pc.OpList[x:]
	vars := Vars{"count": 10, "text": "Hello"}

	vm := NewVM()
	vm.Run(bc, vars, st.output)
	st.Advance()
}

// Executes what's in st.sa
func txFunCallOmni(st *State) {
	t := reflect.ValueOf(st.sa)
	switch t.Kind() {
	case reflect.Int:
		// If it's an int, assume that it's a MACRO, which points to
		// the location in the bytecode that contains the macro code
		txMacroCall(st)
	case reflect.Func:
		txFunCall(st)
	default:
		st.Warnf("Unknown variable as function call: %s\n", st.sa)
		st.sa = nil
		st.Advance()
	}
}
