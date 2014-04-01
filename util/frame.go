package util

// Frame represents a single stack frame. It has a reference to the main
// stack where the actual data resides. Frame is just a convenient
// wrapper to remember when the Frame started
type Frame struct {
  name string
  stack *Stack
  mark int
}

// NewFrame creates a new Frame instance.
func NewFrame(s *Stack) *Frame {
  return &Frame {
    mark: 0,
    stack: s,
  }
}

func (f *Frame) SetMark(v int) {
  f.mark = v
}

// Mark returns the current mark index
func (f *Frame) Mark() int {
  return f.mark
}

// DeclareVar puts a new variable in the stack, and returns the
// index where it now resides
func (f *Frame) DeclareVar(v interface {}) int {
  f.stack.Push(v)
  return f.stack.Cur()
}

// GetLvar gets the frame local variable at position i
func (f *Frame) GetLvar(i int) interface {} {
  v, err := f.stack.Get(i + f.mark)
  if err != nil {
    return nil
  }
  return v
}

// SetLvar sets the frame local variable at position i
func (f *Frame) SetLvar(i int, v interface {}) {
  f.stack.Set(i + f.mark, v)
}

// LastLvarIndex returns the index of the last element in our stack.
func (f *Frame) LastLvarIndex() int {
  return f.stack.Cur()
}


