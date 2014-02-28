package vm

import (
  "bytes"
  "fmt"
)

type OpType int
const (
  TXOP_noop       OpType = iota
  TXOP_nil
  TXOP_literal
  TXOP_fetch_s
  TXOP_print_raw
  TXOP_end
)

var TXCODE_noop = &ExecCode { TXOP_noop, func(st *State) { st.Advance() } }
var TXCODE_end = &ExecCode { TXOP_end, nil }
var TXCODE_literal = &ExecCode { TXOP_literal, func(st *State) {
  st.sa = st.CurrentOp().u_arg
  st.Advance()
} }
var TXCODE_fetch_s = &ExecCode { TXOP_fetch_s, func(st *State) {
  key   := st.CurrentOp().u_arg
  vars  := st.Vars()
  if v, ok := vars.Get(key); ok {
    st.sa = v
  } else {
    st.sa = nil
  }
  st.Advance()
}}
var TXCODE_nil = &ExecCode { TXOP_nil, func(st *State) {
  st.sa = nil
  st.Advance()
} }
var TXCODE_print_raw = &ExecCode { TXOP_print_raw,
  func(st *State) {
    // mark_raw handling
    arg := st.sa
    if arg == nil {
      st.Warnf("Use of nil to print\n")
    } else {
      st.AppendOutput([]byte(fmt.Sprintf("%s", arg)))
    }
    st.Advance()
  },
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

func (ol OpList) String() string {
  buf := &bytes.Buffer {}
  for k, v := range ol {
    buf.WriteString(fmt.Sprintf("%03d. %s\n", k + 1, v))
  }
  return buf.String()
}

