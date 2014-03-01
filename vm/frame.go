package vm

type Frame struct {
  name string
  output interface {} // TODO: what's this?
  retaddr interface {} // TODO: what's this?
  localvars []interface {}
}

func NewFrame() *Frame {
  return &Frame { localvars: make([]interface {}, 0, 10) }
}

func (f *Frame) GetLvar(i int) interface {} {
  if len(f.localvars) <= i {
    return nil
  }
  return f.localvars[i]
}

func (f *Frame) SetLvar(i int, v interface {}) {
  if len(f.localvars) <= i {
    newl := make([]interface{}, i + 1)
    copy(newl, f.localvars)
    f.localvars = newl
  }
  f.localvars[i] = v
}


