package vm

import (
  "bytes"
  "fmt"
  "reflect"
  "strconv"
  "unicode"
  "unicode/utf8"
)

type OpType int
const (
  TXOP_noop       OpType = iota
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
  TXOP_end
)

var TXCODE_noop           = &ExecCode { TXOP_noop, txNoop }
var TXCODE_move_to_sb     = &ExecCode { TXOP_move_to_sb, txMoveToSb }
var TXCODE_move_from_sb   = &ExecCode { TXOP_move_from_sb, txMoveFromSb }
var TXCODE_print_raw      = &ExecCode { TXOP_print_raw, txPrintRaw }
var TXCODE_end            = &ExecCode { TXOP_end, txNoop }
var TXCODE_literal        = &ExecCode { TXOP_literal, txLiteral }
var TXCODE_fetch_s        = &ExecCode { TXOP_fetch_s, txFetchSymbol }
var TXCODE_fetch_field_s  = &ExecCode { TXOP_fetch_field_s, txFetchField }
var TXCODE_save_to_lvar   = &ExecCode { TXOP_save_to_lvar, txSaveToLvar }
var TXCODE_load_lvar      = &ExecCode { TXOP_load_lvar, txLoadLvar }
var TXCODE_nil            = &ExecCode { TXOP_nil, txNil }
var TXCODE_add            = &ExecCode { TXOP_add, txAdd }
var TXCODE_sub            = &ExecCode { TXOP_sub, txSub }
var TXCODE_mul            = &ExecCode { TXOP_mul, txMul }
var TXCODE_div            = &ExecCode { TXOP_div, txDiv }
var TXCODE_and            = &ExecCode { TXOP_and, txAnd }
func convertNumeric(v interface{}) reflect.Value {
  t := reflect.TypeOf(v)
  switch t.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
    return reflect.ValueOf(v)
  default:
    return reflect.ValueOf(0)
  }
}

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

type ExecCode struct {
  id   OpType
  code func(*State)
}

type Op struct {
  code *ExecCode
  u_arg interface {}
}

type OpList []*Op

func (o OpType) String() string {
  var name string
  switch o {
  case TXOP_noop:       name  = "noop"
  case TXOP_nil:        name  = "nil"
  case TXOP_literal:    name  = "literal"
  case TXOP_fetch_s:    name  = "fetch_s"
  case TXOP_print_raw:  name  = "print_raw"
  case TXOP_end:        name  = "end"
  default:              name  = "Unknown"
  }
  return name
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

func (ol *OpList) Get(i int) *Op {
  return (*ol)[i]
}

func (ol *OpList) Append(op *Op) {
  *ol = append(*ol, op)
}

func (ol *OpList) String() string {
  buf := &bytes.Buffer {}
  for k, v := range *ol {
    buf.WriteString(fmt.Sprintf("%03d. %s\n", k + 1, v))
  }
  return buf.String()
}

