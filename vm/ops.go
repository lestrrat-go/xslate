package vm

import (
  "fmt"
  "html"
  "reflect"
  "unicode"
  "unicode/utf8"
  "github.com/lestrrat/go-xslate/functions"
  "github.com/lestrrat/go-xslate/functions/array"
  "github.com/lestrrat/go-xslate/functions/hash"
)

// These TXOP... constants are identifiers for each op
const (
  TXOPNoop OpType = iota
  TXOPNil
  TXOPMoveToSb
  TXOPMoveFromSb
  TXOPLiteral
  TXOPFetchSymbol
  TXOPFetchFieldSymbol
  TXOPMarkRaw
  TXOPUnmarkRaw
  TXOPPrint
  TXOPPrintRaw
  TXOPSaveToLvar
  TXOPLoadLvar
  TXOPAdd
  TXOPSub
  TXOPMul
  TXOPDiv
  TXOPAnd
  TXOPGoto
  TXOPForStart
  TXOPForIter
  TXOPHTMLEscape
  TXOPUriEscape
  TXOPEq
  TXOPNe
  TXOPPopmark
  TXOPPushmark
  TXOPPush
  TXOPPop
  TXOPFunCall
  TXOPFunCallSymbol
  TXOPMethodCall
  TXOPRange
  TXOPMakeArray
  TXOPMakeHash
  TXOPInclude
  TXOPEnd
  TXOPMax
)

var opnames    = make([]string, TXOPMax)
var ophandlers = make([]OpHandler, TXOPMax)
func init () {
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
    case TXOPLiteral:
      h = txLiteral
      n = "literal"
    case TXOPFetchSymbol:
      h = txFetchSymbol
      n = "fetch_s"
    case TXOPFetchFieldSymbol:
      h = txFetchField
      n = "fetch_field_s"
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
    case TXOPEq:
      h = txEq
      n = "eq"
    case TXOPNe:
      h = txNe
      n = "ne"
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
    case TXOPFunCall:
      h = txFunCall
      n = "funcall"
    case TXOPFunCallSymbol:
      h = txFunCallSymbol
      n = "funcall_symbol"
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
    default:
      panic("No such optype")
    }
    ophandlers[i] = h
    opnames[i]    = n
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
  key   := st.CurrentOp().Arg()
  vars  := st.Vars()
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
      f = v.FieldByName(name)
    case reflect.Map:
      v = reflect.ValueOf(container)
      f = v.MapIndex(reflect.ValueOf(name))
    default:
      panic("XXX Put proper error handling here")
    }

    st.sa = f.Interface()
  }
  st.Advance()
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
  idx := st.CurrentOp().ArgInt()
  st.sa = st.CurrentFrame().GetLvar(idx)
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

func txForStart(st *State) {
  id    := st.CurrentOp().ArgInt()
  slice := reflect.ValueOf(st.sa)

  switch slice.Kind() {
  case reflect.Array, reflect.Slice:
    // Normal case. nothing to do
  default:
    // Oh you silly goose. You didn't give me a slice.
    // Use a dummy array
    slice = reflect.ValueOf([]struct{}{})
  }

  cf := st.CurrentFrame()
  cf.SetLvar(id    , nil)   // item
  cf.SetLvar(id + 1, -1)    // index
  cf.SetLvar(id + 2, slice) // slice (Value)

  st.Advance()
}

func txForIter(st *State) {
  id    := st.sa.(int)
  cf    := st.CurrentFrame()
  index := cf.GetLvar(id + 1).(int)
  slice := cf.GetLvar(id + 2).(reflect.Value)

  index++
  cf.SetLvar(id + 1, index)
  if slice.Len() > index {
    cf.SetLvar(id, slice.Index(index).Interface())
    st.Advance()
    return
  }

  // loop done
  st.AdvanceBy(st.CurrentOp().ArgInt())
}

func txUriEscape(st *State) {
  v := interfaceToString(st.sa)
  st.sa = escapeUriString(v)
  st.Advance()
}

func txHTMLEscape(st *State) {
  v := interfaceToString(st.sa)
  st.sa = html.EscapeString(v)
  st.Advance()
}

func txEq(st *State) {
  st.sa = st.sb == st.sa
  st.Advance()
}

func txNe(st *State) {
  st.sa = st.sb != st.sa
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
    st.Warnf("Number of arguments for function does not match (expected %d, got %d)", fun.Type().NumIn(), len(args))
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
  tip  := st.stack.Cur()
  var args []reflect.Value

  if tip - mark - 1  > 0{
    args = make([]reflect.Value, tip - mark - 1)
    for i := mark + 1; tip > i; i++ {
      v, _ := st.stack.Get(i)
      args[i - mark] = reflect.ValueOf(v)
    }
  }

  x := st.sa
  st.sa = nil
  if x != nil {
    v := reflect.ValueOf(x)
    if v.Type().Kind() == reflect.Func {
      fun := reflect.ValueOf(x)
      invokeFuncSingleReturn(st, fun, args)
    }
  }
  st.Advance()
}

func txFunCallSymbol(st *State) {
  // Everything in our lvars up to the current tip is our argument list
  mark := st.CurrentMark()
  tip  := st.stack.Cur()
  var args []reflect.Value

  if tip - mark - 1  > 0{
    args = make([]reflect.Value, tip - mark - 1)
    for i := mark + 1; tip > i; i++ {
      v, _ := st.stack.Get(i)
      args[i - mark] = reflect.ValueOf(v)
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
  name := interfaceToString(st.CurrentOp().Arg())

  // Everything in our lvars up to the current tip is our argument list
  mark := st.CurrentMark()
  tip  := st.stack.Cur()

  v, _ := st.stack.Get(mark)
  invocant := reflect.ValueOf(v)

  var args = make([]reflect.Value, tip - mark + 1)
  for i := 0; tip >= i; i++ {
    v, _ := st.stack.Get(i)
    args[i - mark] = reflect.ValueOf(v)
  }

  // For maps, arrays, slices, we call virtual methods, if they are available
  switch invocant.Kind() {
  case reflect.Map:
    fun, ok := hash.Depot().Get(name)
    if ok {
      invokeFuncSingleReturn(st, fun, args)
    }
  case reflect.Array, reflect.Slice:
    fun, ok := array.Depot().Get(name)
    if ok {
      invokeFuncSingleReturn(st, fun, args)
    }
  default:
    method, ok := invocant.Type().MethodByName(name)
    if ! ok {
      st.sa = nil
    } else {
      invokeFuncSingleReturn(st, method.Func, args)
    }
  }
  st.Advance()
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
  end   := st.StackTip()    // end

  if end <= start {
    panic(fmt.Sprintf("MakeArray: list start (%d) >= end (%d)", start, end))
  }

  list := make([]interface {}, end - start + 1)
  for i := end; i >= start; i-- {
    list[i - start] = st.StackPop()
  }
  st.sa = list
  st.Advance()
}

func txMakeHash(st *State) {
  start := st.CurrentMark() // start
  end   := st.StackTip()    // end

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

  vars := Vars {}
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

  vm := NewVM()
  vm.Run(bc, vars)
  output, err := vm.OutputString()
  if err == nil {
    st.AppendOutputString(output)
  }
  st.Advance()
}
