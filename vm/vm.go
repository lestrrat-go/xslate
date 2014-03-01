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

type Frame struct {
  name string
  output interface {} // TODO: what's this?
  retaddr interface {} // TODO: what's this?
  localvars []interface {}
}

func NewFrame() *Frame {
  return &Frame { localvars: make([]interface {}, 0, 10) }
}

func (f *Frame) GetLvar(i int) interface {} {
  if len(f.localvars) <= i {
    return nil
  }
  return f.localvars[i]
}

func (f *Frame) SetLvar(i int, v interface {}) {
  if len(f.localvars) <= i {
    newl := make([]interface{}, i + 1)
    copy(newl, f.localvars)
    f.localvars = newl
  }
  f.localvars[i] = v
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

  // stack frame
  frames []*Frame // TODO: what's in a frame?
  currentFrame int
}

func NewState() *State {
  return &State {
    opidx: 0,
    pc: &OpList {},
    vars: make(Vars),
    output: &bytes.Buffer {},
    frames: make([]*Frame, 10),
    currentFrame: -1,
  }
}

func (st *State) Advance() {
  st.AdvanceBy(1)
}

func (st *State) AdvanceBy(i int) {
  st.opidx += i
}

func (st *State) Vars() Vars {
  return st.vars
}

func (st *State) CurrentOp() *Op {
  return st.pc.Get(st.opidx)
}

func (st *State) PushFrame(f *Frame) {
  if st.currentFrame >= len(st.frames) {
    newf := make([]*Frame, st.currentFrame + 1)
    copy(newf, st.frames)
    st.frames = newf
  }
  st.currentFrame++
  st.frames[st.currentFrame] = f
}

func (st *State) PopFrame() {
  st.frames[st.currentFrame] = nil
  st.currentFrame--
}

func (st *State) CurrentFrame() *Frame {
  return st.frames[st.currentFrame]
}

func (st *State) Warnf(format string, args ...interface{}) {
  st.warn.Write([]byte(fmt.Sprintf(format, args...)))
}

func (st *State) AppendOutput(b []byte) {
  // XXX Error checking?
  st.output.Write(b)
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
  for op := st.CurrentOp(); op.OpType() != TXOP_end; op = st.CurrentOp() {
    op.Call(st)
  }
}