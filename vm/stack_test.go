package vm

import (
  "testing"
)

type IntWrap struct { i int }

func TestStack_Grow(t *testing.T) {
  s := NewStack(5)
  for i := 0; i < 10; i++ {
    if i % 2 == 0 {
      s.Push(i)
    } else {
      s.Push(IntWrap { i })
    }
  }

  for i := 0; i < 10; i++ {
    x := s.Get(i)
    if i % 2 == 0 {
      if x.(int) != i {
        t.Errorf("Get(%d): Expected %d, got %s\n", i, i, x)
      }
    } else {
      if x.(IntWrap).i != i {
        t.Errorf("Get(%d): Expected %d, got %s\n", i, x.(IntWrap).i, x)
      }
    }
  }

  for i := 9; i > -1; i-- {
    x := s.Pop()
    if i % 2 == 0 {
      if x.(int) != i {
        t.Errorf("Pop(%d): Expected %d, got %s\n", i, i, x)
      }
    } else {
      if x.(IntWrap).i != i {
        t.Errorf("Get(%d): Expected %d, got %s\n", i, x.(IntWrap).i, x)
      }
    }
  }
}

