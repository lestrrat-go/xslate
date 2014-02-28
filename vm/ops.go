package vm

import (
  "bytes"
  "fmt"
  "reflect"
  "strconv"
)

type OpType int
const (
  TXOP_noop       OpType = iota
  TXOP_nil
  TXOP_literal
  TXOP_fetch_s
  TXOP_fetch_field_s
  TXOP_print_raw
  TXOP_end
)

var TXCODE_noop           = &ExecCode { TXOP_noop, txNoop }
var TXCODE_print_raw      = &ExecCode { TXOP_print_raw, txPrintRaw }
var TXCODE_end            = &ExecCode { TXOP_end, txNoop }
var TXCODE_literal        = &ExecCode { TXOP_literal, txLiteral }
var TXCODE_fetch_s        = &ExecCode { TXOP_fetch_s, txFetchSymbol }
var TXCODE_fetch_field_s  = &ExecCode { TXOP_fetch_field_s, txFetchField }
var TXCODE_nil            = &ExecCode { TXOP_nil, txNil }

func txNil(st *State) {
  st.sa = nil
  st.Advance()
}

func txNoop(st *State) {
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
  } else {
    t := reflect.TypeOf(container)
    var v reflect.Value
    switch t.Kind() {
    case reflect.Ptr, reflect.Struct:
      v = reflect.ValueOf(container)
    default:
      v = reflect.ValueOf(&container).Elem()
    }
    name := fmt.Sprintf("%s", st.CurrentOp().u_arg)
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
    t := reflect.TypeOf(arg)
    var v string
    switch t.Kind() {
    case reflect.String:
      v, _ = arg.(string)
    case reflect.Int:
      x, _ := arg.(int)
      v = strconv.FormatInt(int64(x), 10)
    case reflect.Int64:
      x, _ := arg.(int64)
      v = strconv.FormatInt(x, 10)
    case reflect.Int32:
      x, _ := arg.(int32)
      v = strconv.FormatInt(int64(x), 10)
    case reflect.Int16:
      x, _ := arg.(int16)
      v = strconv.FormatInt(int64(x), 10)
    case reflect.Int8:
      x, _ := arg.(int8)
      v = strconv.FormatInt(int64(x), 10)
    default:
      v = fmt.Sprintf("%s", arg)
    }
    st.AppendOutput([]byte(v))
  }
  st.Advance()
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

