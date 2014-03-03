package vm

import (
  "testing"
)

func TestStack_Grow(t *testing.T) {
  s := NewStack(5)
  for i := 0; i < 10; i++ {
    s.Push(i)
  }

  for i := 0; i < 10; i++ {
    x := s.Get(i)
    if x.(int) != i {
      t.Errorf("Get(%d): Expected %d, got %s\n", i, i, x)
    }
  }

  for i := 9; i > -1; i-- {
    x := s.Pop()
    if x.(int) != i {
      t.Errorf("Pop(): Expected %d, got %s\n", i, x)
    }
  }
}

