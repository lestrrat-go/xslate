package vm

import (
  "bytes"
  "fmt"
  "io"
  "os"
)

type State struct {
  opidx int
  pc *OpList

  stack *Stack
  markstack *Stack

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
  frames *Stack
  currentFrame int
}

func NewState() *State {
  st := &State {
    opidx: 0,
    pc: &OpList {},
    stack: NewStack(5),
    markstack: NewStack(5),
    vars: make(Vars),
    output: &bytes.Buffer {},
    warn: os.Stderr,
    frames: NewStack(5),
    currentFrame: -1,
  }

  st.Pushmark()
  st.PushFrame(NewFrame(st.CurrentMark(), st.stack))
  return st
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
  st.frames.Push(f)
}

func (st *State) PopFrame() *Frame {
  x := st.frames.Pop()
  return x.(*Frame)
}

func (st *State) CurrentFrame() *Frame {
  x := st.frames.Top()
  return x.(*Frame)
}

func (st *State) Warnf(format string, args ...interface{}) {
  st.warn.Write([]byte(fmt.Sprintf(format, args...)))
}

func (st *State) AppendOutput(b []byte) {
  // XXX Error checking?
  st.output.Write(b)
}

func (st *State) AppendOutputString(o string) {
  st.output.Write([]byte(o))
}

func (st *State) Pushmark() {
  st.markstack.Push(st.stack.Cur())
}

func (st *State) Popmark() int {
  x := st.markstack.Pop()
  return x.(int)
}

func (st *State) CurrentMark() int {
  x := st.markstack.Top()
  return x.(int)
}

func (st *State) StackTip() int {
  return st.stack.Cur()
}
