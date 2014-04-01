package util

import (
  "bytes"
  "fmt"
)

// Stack is a simple structure to hold various data
type Stack struct {
  cur  int
  data []interface {}
}

func calcNewSize(base int) int {
  return int(float64(base) * 1.5)
}

// NewStack creates a new Stack of initial size `size`.
func NewStack(size int) *Stack {
  return &Stack {
    cur: -1,
    data: make([]interface {}, size),
  }
}

// Cur returns the current stack index. Note that if nothing is stored in
// the stack, it returns -1
func (s *Stack) Cur() int {
  return s.cur
}

// SetCur sets the current cursor location
func (s *Stack) SetCur(c int) {
  s.cur = c
}

// Top returns the element at the top of the stack or an error if stack is empty
func (s *Stack) Top() (interface {}, error) {
  cur := s.Cur()
  if cur < 0 {
    return nil, fmt.Errorf("error: nothing in stack")
  }
  v, err :=  s.Get(cur)
  if err != nil {
    return nil, err
  }

  return v, nil
}
// BufferSize returns the length of the underlying buffer
func (s *Stack) BufferSize() int {
  return len(s.data)
}

// Size returns the number of elements stored in this stack
func (s *Stack) Size() int {
  return s.cur + 1
}

// Resize changes the size of the underlying buffer
func (s *Stack) Resize(size int) {
  newl := make([]interface {}, size)
  copy(newl, s.data)
  s.data = newl
}

// Extend changes the size of the underlying buffer, extending it by `extendBy`
func (s *Stack) Extend(extendBy int) {
  s.Resize(s.Size() + extendBy)
}

// Grow automatically grows the underlying buffer so that it can hold at
// least `min` elements
func (s *Stack) Grow(min int) {
  // Automatically grow the stack to some long-enough length
  if min <= s.BufferSize() {
    // we have enough
    return
  }

  s.Resize(calcNewSize(min))
}

// Get returns the element at position `i`
func (s *Stack) Get(i int) (interface {}, error) {
  if i < 0 || i >= len(s.data) {
    return nil, fmt.Errorf("%d is out of range", i)
  }

  d := s.data
  return d[i], nil
}

// Set sets the element at position `i` to `v`. The stack size is automatically
// adjusted.
func (s *Stack) Set(i int, v interface {}) {
  if i >= s.BufferSize() {
    s.Resize(calcNewSize(i))
  }

  d := s.data
  d[i] = v
}

// Push adds an element at the end of the stack
func (s *Stack) Push(v interface {}) {
  s.cur++
  if s.cur >= s.BufferSize() {
    s.Resize(calcNewSize(s.cur))
  }
  s.data[s.cur] = v
}

// Pop removes and returns the item at the end of the stack
func (s *Stack) Pop() interface {} {
  if s.cur < 0 {
    return nil
  }

  v := s.data[s.cur]
  s.data[s.cur] = nil
  s.cur--
  return v
}

// String returns the textual representation of the stack
func (s *Stack) String() string {
  buf := &bytes.Buffer {}
  for k, v := range s.data {
    mark := " "
    if k == s.cur {
      mark = "*"
    }
    fmt.Fprintf(buf, "%s %03d: %q\n", mark, k, v)
  }
  return buf.String()
}
