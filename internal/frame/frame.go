package frame

import (
	"strconv"

	"github.com/lestrrat-go/xslate/internal/stack"
	"github.com/pkg/errors"
)

// Frame represents a single stack frame. It has a reference to the main
// stack where the actual data resides. Frame is just a convenient
// wrapper to remember when the Frame started
type Frame struct {
	stack stack.Stack
	mark  int
}

// New creates a new Frame instance.
func New(s stack.Stack) *Frame {
	return &Frame{
		mark:  0,
		stack: s,
	}
}

func (f Frame) Stack() stack.Stack {
	return f.stack
}

// SetMark sets the offset from which this frame's variables may be stored
func (f *Frame) SetMark(v int) {
	f.mark = v
}

// Mark returns the current mark index
func (f *Frame) Mark() int {
	return f.mark
}

// DeclareVar puts a new variable in the stack, and returns the
// index where it now resides
func (f *Frame) DeclareVar(v interface{}) int {
	f.stack.Push(v)
	return f.stack.Size() - 1
}

// GetLvar gets the frame local variable at position i
func (f *Frame) GetLvar(i int) (interface{}, error) {
	v, err := f.stack.Get(i)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get local variable at "+strconv.Itoa(i+f.mark))
	}
	return v, nil
}

// SetLvar sets the frame local variable at position i
func (f *Frame) SetLvar(i int, v interface{}) {
	f.stack.Set(i, v)
}

// LastLvarIndex returns the index of the last element in our stack.
func (f *Frame) LastLvarIndex() int {
	return f.stack.Size()
}
