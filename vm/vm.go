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

type Vars map[string]interface {}

func (v Vars) Set(k string, x interface {}) {
  v[k] = x
}

func (v Vars) Get(k interface {}) (interface{}, bool) {
  key := fmt.Sprintf("%s", k)
  x, ok := v[key]
  return x, ok
}

type State struct {
  opidx int
  pc *OpList

  // output
  output  io.ReadWriter
  warn    io.Writer

  // template variables
  vars Vars

  // registers
  sa    interface {}
  sb    interface {}
  targ  interface {}
}

func (st *State) Advance() {
  st.opidx++
}

func (st *State) Vars() Vars {
  return st.vars
}

func (st *State) CurrentOp() *Op {
  return st.pc.Get(st.opidx)
}

func (st *State) Warnf(format string, args ...interface{}) {
  st.warn.Write([]byte(fmt.Sprintf(format, args...)))
}

func (st *State) AppendOutput(b []byte) {
  // XXX Error checking?
  st.output.Write(b)
}

func NewVM() (*VM) {
  return &VM {
    &State {
      opidx: 0,
      pc: &OpList {},
      vars: make(Vars),
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