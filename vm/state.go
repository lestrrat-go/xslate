package vm

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "os"

  "github.com/lestrrat/go-xslate/util"
)

// State keeps track of Xslate Virtual Machine state
type State struct {
  opidx int
  pc *ByteCode

  stack *util.Stack
  markstack *util.Stack

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
  frames *util.Stack
  currentFrame int

  Loader byteCodeLoader
}

// NewState creates a new State struct
func NewState() *State {
  st := &State {
    opidx: 0,
    pc: NewByteCode(),
    stack: util.NewStack(5),
    markstack: util.NewStack(5),
    vars: make(Vars),
    output: &bytes.Buffer {},
    warn: os.Stderr,
    frames: util.NewStack(5),
    currentFrame: -1,
  }

  st.Pushmark()
  st.PushFrame(NewFrame())
  return st
}

// Advance advances the op code position by 1 
func (st *State) Advance() {
  st.AdvanceBy(1)
}

// AdvanceBy advances the op code position by `i`
func (st *State) AdvanceBy(i int) {
  st.opidx += i
}

// Vars returns the current set of variables
func (st *State) Vars() Vars {
  return st.vars
}

// CurrentOp returns the current op code
func (st *State) CurrentOp() *Op {
  return st.pc.Get(st.opidx)
}

// PushFrame pushes a new frame to the frame stack
func (st *State) PushFrame(f *Frame) {
  st.frames.Push(f)
}

// PopFrame pops the frame from the top of the frame stack
func (st *State) PopFrame() *Frame {
  x := st.frames.Pop()
  if x == nil {
    return nil
  }
  return x.(*Frame)
}

// CurrentFrame returns the frame currently at the top of the frame stack
func (st *State) CurrentFrame() *Frame {
  x, err := st.frames.Top()
  if err != nil {
    return nil
  }
  return x.(*Frame)
}

// Warnf is used to generate warnings during virtual machine execution
func (st *State) Warnf(format string, args ...interface{}) {
  st.warn.Write([]byte(fmt.Sprintf(format, args...)))
}

// AppendOutput appends the specified bytes to the output
func (st *State) AppendOutput(b []byte) {
  // XXX Error checking?
  st.output.Write(b)
}

// AppendOutputString is the same as AppendOutput, but uses a string
func (st *State) AppendOutputString(o string) {
  st.output.Write([]byte(o))
}

// Output returns the accumulated output
func (st *State) Output() ([]byte, error) {
  return ioutil.ReadAll(st.output)
}

// OutputString returns the accumulated output as string
func (st *State) OutputString() (string, error) {
  buf, err := st.Output()
  if err != nil {
    return "", err
  }
  return string(buf), nil
}

// Pushmark records the current stack tip so we can remember
// where the current context started
func (st *State) Pushmark() {
  cur := st.stack.Cur()
  if cur < 0 {
    cur = 0
  }
  st.markstack.Push(cur)
}

// Popmark pops the mark stored at the top of the mark stack
func (st *State) Popmark() int {
  x := st.markstack.Pop()
  return x.(int)
}

// CurrentMark returns the mark stored at the top of the mark stack
func (st *State) CurrentMark() int {
  x, err := st.markstack.Top()
  if err != nil {
    x = 0
  }
  return x.(int)
}

// StackTip returns the index of the top of the stack
func (st *State) StackTip() int {
  return st.stack.Cur()
}

// StackPop pops from the stack
func (st *State) StackPop() interface {} {
  return st.stack.Pop()
}

// StackPush pushes to the stack
func (st *State) StackPush(v interface {}) {
  st.stack.Push(v)
}

// LoadByteCode loads a new ByteCode. This is used for op codes that
// call to external templates such as `include`
func (st *State) LoadByteCode(key string) (*ByteCode, error) {
  return st.Loader.Load(key)
}
