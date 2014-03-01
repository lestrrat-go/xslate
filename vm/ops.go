package vm

import (
  "encoding/json"
  "fmt"
  "reflect"
  "strconv"
  "unicode"
  "unicode/utf8"
)

type OpType int
type OpHandler func(*State)
type ExecCode struct {
  id   OpType
  code func(*State)
}

const (
  TXOP_noop OpType = iota
  TXOP_nil
  TXOP_move_to_sb
  TXOP_move_from_sb
  TXOP_literal
  TXOP_fetch_s
  TXOP_fetch_field_s
  TXOP_print_raw
  TXOP_save_to_lvar
  TXOP_load_lvar
  TXOP_add
  TXOP_sub
  TXOP_mul
  TXOP_div
  TXOP_and
  TXOP_goto
  TXOP_end
  TXOP_max
)

var opnames    []string    = make([]string, TXOP_max)
var ophandlers []OpHandler = make([]OpHandler, TXOP_max)
var execcodes  []*ExecCode = make([]*ExecCode, TXOP_max)
func init () {
  for i := TXOP_noop; i < TXOP_max; i++ {
    var h OpHandler
    n := "Unknown"
    switch i {
    case TXOP_noop:
      h = txNoop
      n = "noop"
    case TXOP_end:
      h = txEnd
      n = "end"
    case TXOP_move_to_sb:
      h = txMoveToSb
      n = "move_to_sb"
    case TXOP_move_from_sb:
      h = txMoveFromSb
      n = "move_from_sb"
    case TXOP_print_raw:
      h = txPrintRaw
      n = "print_raw"
    case TXOP_literal:
      h = txLiteral
      n = "literal"
    case TXOP_fetch_s:
      h = txFetchSymbol
      n = "fetch_s"
    case TXOP_fetch_field_s:
      h = txFetchField
      n = "fetch_field_s"
    case TXOP_save_to_lvar:
      h = txSaveToLvar
      n = "save_to_lvar"
    case TXOP_load_lvar:
      h = txLoadLvar
      n = "load_lvar"
    case TXOP_nil:
      h = txNil
      n = "nil"
    case TXOP_add:
      h = txAdd
      n = "add"
    case TXOP_sub:
      h = txSub
      n = "sub"
    case TXOP_mul:
      h = txMul
      n = "mul"
    case TXOP_div:
      h = txDiv
      n = "div"
    case TXOP_and:
      h = txAnd
      n = "and"
    case TXOP_goto:
      h = txGoto
      n = "goto"
    default:
      panic("No such optype")
    }
    ophandlers[i] = h
    execcodes[i]  = &ExecCode { OpType(i), h}
    opnames[i]    = n
  }
}

func optypeToExecCode(o OpType) *ExecCode {
  return execcodes[o]
}

func optypeToHandler(o OpType) OpHandler {
  return ophandlers[o]
}

func convertNumeric(v interface{}) reflect.Value {
  t := reflect.TypeOf(v)
  switch t.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
    return reflect.ValueOf(v)
  default:
    return reflect.ValueOf(0)
  }
}

// Given possibly non-matched pair of things to perform arithmetic
// operations on, align their types so that the given operation
// can be performed correctly.
// e.g. given int, float, we align them to float, float
func alignTypesForArithmetic(left, right interface {}) (reflect.Value, reflect.Value) {
  leftV  := convertNumeric(left)
  rightV := convertNumeric(right)

  if leftV.Kind() == rightV.Kind() {
    return leftV, rightV
  }

  var alignTo reflect.Type
  if leftV.Kind() > rightV.Kind() {
    alignTo = leftV.Type()
  } else {
    alignTo = rightV.Type()
  }

  return leftV.Convert(alignTo), rightV.Convert(alignTo)
}

func interfaceToString(arg interface {}) string {
  t := reflect.TypeOf(arg)
  var v string
  switch t.Kind() {
  case reflect.String:
    v, _ = arg.(string)
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    v = strconv.FormatInt(reflect.ValueOf(arg).Int(), 10)
  case reflect.Float32, reflect.Float64:
    v = strconv.FormatFloat(reflect.ValueOf(arg).Float(), 'f', -1, 64)
  default:
    v = fmt.Sprintf("%s", arg)
  }
  return v
}

func interfaceToBool(arg interface {}) bool {
  t := reflect.TypeOf(arg)
  if t.Kind() == reflect.Bool {
    return arg.(bool)
  }

  z := reflect.Zero(t)
  return reflect.DeepEqual(z, t)
}

func txEnd(st *State) {}

func txNil(st *State) {
  st.sa = nil
  st.Advance()
}

func txNoop(st *State) {
  st.Advance()
}

func txMoveToSb(st *State) {
  st.sb = st.sa
  st.Advance()
}

func txMoveFromSb(st *State) {
  st.sa = st.sb
  st.Advance()
}

func txLiteral(st *State) {
  st.sa = st.CurrentOp().u_arg
  st.Advance()
}

func txFetchSymbol(st *State) {
  key   := st.CurrentOp().u_arg
  vars  := st.Vars()
  if v, ok := vars.Get(key); ok {
    st.sa = v
  } else {
    st.sa = nil
  }
  st.Advance()
}

func txFetchField(st *State) {
  container := st.sa
  if container == nil {
    // XXX ? no op?
    st.sa = nil
  } else {
    t := reflect.TypeOf(container)
    var v reflect.Value
    switch t.Kind() {
    case reflect.Ptr, reflect.Struct:
      v = reflect.ValueOf(container)
    default:
      v = reflect.ValueOf(&container).Elem()
    }
    name := interfaceToString(st.CurrentOp().u_arg)
    r, size := utf8.DecodeRuneInString(name)
    name = string(unicode.ToUpper(r)) + name[size:]
    f := v.FieldByName(name)
    st.sa = f.Interface()
  }
  st.Advance()
}

func txPrintRaw(st *State) {
  // mark_raw handling
  arg := st.sa
  if arg == nil {
    st.Warnf("Use of nil to print\n")
  } else {
    v := interfaceToString(arg)
    st.AppendOutput([]byte(v))
  }
  st.Advance()
}

func txSaveToLvar(st *State) {
  idx, ok := st.CurrentOp().u_arg.(int)
  if ! ok {
    panic("save_to_lvar.u_arg MUST BE AN INT")
  }

  st.CurrentFrame().SetLvar(idx, st.sa)
  st.Advance()
}

func txLoadLvar(st *State) {
  idx, ok := st.CurrentOp().u_arg.(int)
  if ! ok {
    panic("load_lvar.u_arg MUST BE AN INT")
  }

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
    st.AdvanceBy(int(reflect.ValueOf(st.CurrentOp().u_arg).Int()))
  }
}

func txGoto(st *State) {
  st.AdvanceBy(int(reflect.ValueOf(st.CurrentOp().u_arg).Int()))
}

type Op struct {
  code  *ExecCode
  u_arg interface {}
}

func NewOp(o OpType, args ...interface {}) *Op {
  e := optypeToExecCode(o)
  var arg interface {}
  if len(args) > 0 {
    arg = args[0]
  } else {
    arg = nil
  }
  return &Op { e, arg }
}

func (o Op) MarshalJSON() ([]byte, error) {
  return json.Marshal(map[string]interface{}{
    "code": o.code.id.String(),
    "u_arg": o.u_arg,
  })
}

func (o OpType) String() string {
  return opnames[o]
}

func (o *Op) Call(st *State) {
  o.code.code(st)
}
func (o *Op) OpType() OpType {
  return o.code.id
}
func (o *Op) Code() *ExecCode {
  return o.code
}

func (o *Op) String() string {
  // TODO: also print out register id's and stuff
  return fmt.Sprintf("Op[%s]", o.OpType())
}

