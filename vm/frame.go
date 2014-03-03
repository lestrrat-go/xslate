package vm

type Frame struct {
  name string

  stack *Stack
  base int // base index into the main stack

  output interface {} // TODO: what's this?
  retaddr interface {} // TODO: what's this?
}

// a Frame uses the stack, starting at position `base`. The base is determined
// by calling st.Pushmark() before calling NewFrame()
func NewFrame(base int, s *Stack) *Frame {
  return &Frame {
    base: base,
    stack: s,
  }
}

// Gets the frame local variable at position i, which is relative from base
func (f *Frame) GetLvar(i int) interface {} {
  return f.stack.Get(f.base + i)
}

// Sets the frame local variable at position i, which is relative from base
func (f *Frame) SetLvar(i int, v interface {}) {
  f.stack.Set(f.base + i, v)
}

// Returns the index of the last element in our stack, relative from base
func (f *Frame) LastLvarIndex() int {
  return f.stack.Cur() - f.base
}


