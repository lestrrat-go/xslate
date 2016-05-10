package vm

import (
	"fmt"
	"os"

	"github.com/lestrrat/go-xslate/internal/frame"
	"github.com/lestrrat/go-xslate/internal/stack"
)

// NewState creates a new State struct
func NewState() *State {
	st := &State{
		opidx:      0,
		pc:         NewByteCode(),
		stack:      stack.New(5),
		markstack:  stack.New(5),
		framestack: stack.New(5),
		frames:     stack.New(5),
		vars:       make(Vars),
		warn:       os.Stderr,
		MaxLoopCount: 1000,
	}

	st.Pushmark()
	st.PushFrame()
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

// AdvanceTo advances the op code to exactly position `i`
func (st *State) AdvanceTo(i int) {
	st.opidx = i
}

// CurrentPos returns the position of the current executing op
func (st *State) CurrentPos() int {
	return st.opidx
}

// Vars returns the current set of variables
func (st *State) Vars() Vars {
	return st.vars
}

// CurrentOp returns the current op code
func (st *State) CurrentOp() Op {
	return st.pc.Get(st.opidx)
}

// PushFrame pushes a new frame to the frame stack
func (st *State) PushFrame() *frame.Frame {
	f := frame.New(st.framestack)
	st.frames.Push(f)
	f.SetMark(st.frames.Size())
	return f
}

// PopFrame pops the frame from the top of the frame stack
func (st *State) PopFrame() *frame.Frame {
	x := st.frames.Pop()
	if x == nil {
		return nil
	}
	f := x.(*frame.Frame)
	for i := st.framestack.Size(); i > f.Mark(); i-- {
		st.framestack.Pop()
	}
	return f
}

// CurrentFrame returns the frame currently at the top of the frame stack
func (st *State) CurrentFrame() *frame.Frame {
	x, err := st.frames.Top()
	if err != nil {
		return nil
	}
	return x.(*frame.Frame)
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

// Pushmark records the current stack tip so we can remember
// where the current context started
func (st *State) Pushmark() {
	st.markstack.Push(st.stack.Size())
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
	return st.stack.Size()-1
}

// StackPop pops from the stack
func (st *State) StackPop() interface{} {
	return st.stack.Pop()
}

// StackPush pushes to the stack
func (st *State) StackPush(v interface{}) {
	st.stack.Push(v)
}

// LoadByteCode loads a new ByteCode. This is used for op codes that
// call to external templates such as `include`
func (st *State) LoadByteCode(key string) (*ByteCode, error) {
	return st.Loader.Load(key)
}

// Reset resets the whole State object
func (st *State) Reset() {
	st.opidx = 0
	st.sa = nil
	st.sb = nil
	st.stack.Reset()
	st.markstack.Reset()
	st.frames.Reset()
	st.framestack.Reset()

	st.Pushmark()
	st.PushFrame()
}
