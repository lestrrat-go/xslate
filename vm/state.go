package vm

import (
  "bytes"
  "io"
  "fmt"
)

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
  st := &State {
    opidx: 0,
    pc: &OpList {},
    vars: make(Vars),
    output: &bytes.Buffer {},
    frames: make([]*Frame, 10),
    currentFrame: -1,
  }
  st.PushFrame(NewFrame())
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

func (st *State) AppendOutputString(o string) {
  st.output.Write([]byte(o))
}

