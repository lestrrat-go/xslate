package vm

import (
  "bytes"
  "io"
  "io/ioutil"
  "fmt"
)

type VM struct {
  st *State
}


type State struct {
  opidx int
  pc []*Op

  // output
  output io.ReadWriter

  // registers
  sa    interface {}
  sb    interface {}
  targ  interface {}
}

func (st *State) Advance() {
  st.opidx++
}

func (st *State) CurrentOp() *Op {
  return st.pc[st.opidx]
}

func (st *State) Warnf(format string, args ...interface{}) {
  fmt.Printf(format, args...)
}

func (st *State) AppendOutput(b []byte) {
  // XXX Error checking?
  st.output.Write(b)
}

func NewVM() (*VM) {
  return &VM {
    &State {
      opidx: 0,
      pc: []*Op{},
      output: &bytes.Buffer {},
    },
  }
}

func (vm *VM) CurrentOp() *Op {
  return vm.st.CurrentOp()
}

func (vm *VM) Output() ([]byte, error) {
  return ioutil.ReadAll(vm.st.output)
}

func (vm *VM) OutputString() (string, error) {
  buf, err := vm.Output()
  if err != nil {
    return "", err
  }
  return string(buf), nil
}

func (vm *VM) Run() {
  st := vm.st
  for op := st.CurrentOp(); op.OpType() != TXOP_end; op = st.CurrentOp() {
    op.Call(st)
  }
}