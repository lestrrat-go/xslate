package vm

import "fmt"

type OpType int
const (
  TXOP_noop       OpType = iota
  TXOP_nil
  TXOP_literal
  TXOP_print_raw
  TXOP_end
)

var TXCODE_noop = &ExecCode { TXOP_noop, func(st *State) { st.Advance() } }
var TXCODE_end = &ExecCode { TXOP_end, nil }
var TXCODE_literal = &ExecCode { TXOP_literal, func(st *State) {
  st.sa = st.CurrentOp().u_arg
  st.Advance()
} }
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
  return fmt.Sprintf("Op[%s]", o.OpType())
}
