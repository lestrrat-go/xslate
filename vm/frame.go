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

func (f *Frame) ReallocLvars(size int) {
  newl := make([]interface{}, size)
  copy(newl, f.localvars)
  f.localvars = newl
}

func (f *Frame) ExtendLvars() {
  extendTo := int(float64(len(f.localvars)) * 1.5)
  if extendTo == 0 {
    extendTo = 5
  }

  f.ReallocLvars(extendTo)
}

func (f *Frame) SetLvar(i int, v interface {}) {
  if len(f.localvars) <= i {
    f.ExtendLvars()
  }
  f.localvars[i] = v
}


