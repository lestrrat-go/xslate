package vm

import (
  "bytes"
  "fmt"
)
type Stack struct {
  cur  int
  data []interface {}
}

func calcNewSize(base int) int {
  return int(float64(base) * 1.5)
}

func NewStack(size int) *Stack {
  return &Stack {
    cur: -1,
    data: make([]interface {}, size),
  }
}

func (s *Stack) Cur() int {
  if s.cur < 0 {
    return 0
  }
  return s.cur
}

func (s *Stack) Top() interface {} {
  return s.Get(s.Cur())
}

func (s *Stack) BufferSize() int {
  return len(s.data)
}

func (s *Stack) Size() int {
  return s.cur + 1
}

func (s *Stack) Resize(size int) {
  newl := make([]interface {}, size)
  copy(newl, s.data)
  s.data = newl
}

func (s *Stack) Extend(extendBy int) {
  s.Resize(s.Size() + extendBy)
}

func (s *Stack) Grow(min int) {
  // Automatically grow the stack to some long-enough length
  if min <= s.BufferSize() {
    // we have enough
    return
  }

  s.Resize(calcNewSize(min))
}

func (s *Stack) Get(i int) interface {} {
  d := s.data
  return d[i]
}

func (s *Stack) Set(i int, v interface {}) {
  if i >= s.BufferSize() {
    s.Resize(calcNewSize(i))
  }

  d := s.data
  d[i] = v
}

func (s *Stack) Push(v interface {}) {
  s.cur++
  if s.cur >= s.BufferSize() {
    s.Resize(calcNewSize(s.cur))
  }
  s.data[s.cur] = v
}

func (s *Stack) Pop() interface {} {
  v := s.data[s.cur]
  s.data[s.cur] = nil
  s.cur--
  return v
}

func (s *Stack) String() string {
  buf := &bytes.Buffer {}
  for k, v := range s.data {
    mark := " "
    if k == s.cur {
      mark = "*"
    }
    fmt.Fprintf(buf, "%s %03d: %s\n", mark, k, v)
  }
  return buf.String()
}
