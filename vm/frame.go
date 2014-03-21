package vm

// Frame represents a stack frame
type Frame struct {
  name string
  stack *Stack
}

// NewFrame creates a new Frame instance.
func NewFrame() *Frame {
  return &Frame {
    stack: NewStack(5),
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
  if i > f.stack.cur {
    f.stack.cur = i
  }
}

// LastLvarIndex returns the index of the last element in our stack.
func (f *Frame) LastLvarIndex() int {
  return f.stack.Cur()
}


