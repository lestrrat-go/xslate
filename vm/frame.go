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
func NewFrame() *Frame {
  return &Frame {
    stack: NewStack(5),
  }
}

// Gets the frame local variable at position i, which is relative from base
func (f *Frame) GetLvar(i int) interface {} {
  return f.stack.Get(i)
}

// Sets the frame local variable at position i, which is relative from base
func (f *Frame) SetLvar(i int, v interface {}) {
  f.stack.Set(i, v)
  if i > f.stack.cur {
    f.stack.cur = i
  }
}

// Returns the index of the last element in our stack, relative from base
func (f *Frame) LastLvarIndex() int {
  return f.stack.Cur()
}


