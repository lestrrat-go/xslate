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

func (vm *VM) Reset() {
  vm.st.opidx = 0
  vm.st.output = &bytes.Buffer {}
}

func (vm *VM) Run(bc *ByteCode) {
  vm.Reset()
  st := vm.st
  if bc != nil {
    st.pc = bc
  }
  for op := st.CurrentOp(); op.OpType() != TXOP_end; op = st.CurrentOp() {
    op.Call(st)
  }
}