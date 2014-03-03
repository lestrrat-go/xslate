package vm

type Frame struct {
  name string

  stack *Stack
  base int // base index into the main stack

  output interface {} // TODO: what's this?
  retaddr interface {} // TODO: what's this?
}

func NewFrame(base int, s *Stack) *Frame {
  return &Frame {
    base: base,
    stack: s,
  }
}

func (f *Frame) GetLvar(i int) interface {} {
  return f.stack.Get(f.base + i)
}

func (f *Frame) SetLvar(i int, v interface {}) {
  f.stack.Set(f.base + i, v)
}


