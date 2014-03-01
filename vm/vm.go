package vm

import (
  "bytes"
  "io/ioutil"
)

type VM struct {
  st *State
}

func NewVM() (*VM) {
  return &VM { NewState() }
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
  st.opidx = 0
  st.output = &bytes.Buffer {}
  for op := st.CurrentOp(); op.OpType() != TXOP_end; op = st.CurrentOp() {
    op.Call(st)
  }
}