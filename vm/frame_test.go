package vm

import (
  "testing"
)

func TestFrame_Lvar(t *testing.T) {
  f := NewFrame()
  f.SetLvar(0, 1)
  x := f.GetLvar(0)
  i, ok := x.(int)
  if !ok {
    t.Errorf("GetLvar(0) did not return an int")
  } else {
    if i != 1 {
      t.Errorf("GetLvar(0) did not return 1, it returned %d", i)
    }
  }
}

func TestFrame_LvarExtend(t *testing.T) {
  f := NewFrame()

  for i := 0; i < 100; i++ {
    f.SetLvar(i, i)
  }

  if len(f.localvars) != 100 {
    t.Errorf("Expected 100 localvars, but got %d", len(f.localvars))
  }

  for i := 0; i < 100; i++ {
    x := f.GetLvar(i)
    v, ok := x.(int)
    if ! ok {
      t.Errorf("var(%d) is not an int!", i)
    } else if v != i {
      t.Errorf("expected %d, got %d", i, v)
    }
  }
}