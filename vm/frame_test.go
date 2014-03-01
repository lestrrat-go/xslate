package vm

import (
  "testing"
)

func TestLvar(t *testing.T) {
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

func TestLvarExtend(t *testing.T) {
  f := NewFrame()

  for i := 0; i < 100; i++ {
    f.SetLvar(i, i)
  }

  if len(f.localvars) != 100 {
    t.Errorf("Expected 100 localvars, but got %d", len(f.localvars))
  }
}