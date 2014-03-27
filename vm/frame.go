package vm

import (
  "github.com/lestrrat/go-xslate/util"
)

// Frame represents a stack frame
type Frame struct {
  name string
  stack *util.Stack
}

// NewFrame creates a new Frame instance.
func NewFrame() *Frame {
  return &Frame {
    stack: util.NewStack(5),
  }
}

// GetLvar gets the frame local variable at position i
func (f *Frame) GetLvar(i int) interface {} {
  v, err := f.stack.Get(i)
  if err != nil {
    return nil
  }
  return v
}

// SetLvar sets the frame local variable at position i
func (f *Frame) SetLvar(i int, v interface {}) {
  f.stack.Set(i, v)
  if i > f.stack.Cur() {
    f.stack.SetCur(i)
  }
}

// LastLvarIndex returns the index of the last element in our stack.
func (f *Frame) LastLvarIndex() int {
  return f.stack.Cur()
}


