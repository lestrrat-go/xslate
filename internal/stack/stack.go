package stack

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Stack is a simple structure to hold various data
type Stack []interface{}

// Reset clears the contents of the stack and pushes back the cursor
// as if nothing is in the stack
func (s *Stack) Reset() {
	*s = (*s)[:0]
}

func calcNewSize(base int) int {
	if base < 100 {
		return base * 2
	}
	return int(float64(base) * 1.5)
}

// NewStack creates a new Stack of initial size `size`.
func New(size int) Stack {
	return Stack(make([]interface{}, 0, size))
}

// Top returns the element at the top of the stack or an error if stack is empty
func (s *Stack) Top() (interface{}, error) {
	if len(*s) == 0 {
		return nil, errors.New("nothing on stack")
	}
	return (*s)[len(*s)-1], nil
}

// BufferSize returns the length of the underlying buffer
func (s *Stack) BufferSize() int {
	return cap(*s)
}

// Size returns the number of elements stored in this stack
func (s *Stack) Size() int {
	return len(*s)
}

// Resize changes the size of the underlying buffer
func (s *Stack) Resize(size int) {
	newl := make([]interface{}, len(*s), size)
	copy(newl, *s)
	*s = newl
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
func (s *Stack) Get(i int) (interface{}, error) {
	if i < 0 || i >= len(*s) {
		return nil, errors.New(strconv.Itoa(i) + " is out of range")
	}

	return (*s)[i], nil
}

// Set sets the element at position `i` to `v`. The stack size is automatically
// adjusted.
func (s *Stack) Set(i int, v interface{}) error {
	if i < 0 {
		return errors.New("invalid index into stack")
	}

	if i >= s.BufferSize() {
		s.Resize(calcNewSize(i))
	}

	for len(*s) < i + 1 {
		*s = append(*s, nil)
	}

	(*s)[i] = v
	return nil
}

// Push adds an element at the end of the stack
func (s *Stack) Push(v interface{}) {
	if len(*s) >= s.BufferSize() {
		s.Resize(calcNewSize(cap(*s)))
	}

	*s = append(*s, v)
}

// Pop removes and returns the item at the end of the stack
func (s *Stack) Pop() interface{} {
	l := len(*s)
	if l == 0 {
		return nil
	}

	v := (*s)[l-1]
	*s = (*s)[:l-1]
	return v
}

// String returns the textual representation of the stack
func (s *Stack) String() string {
	buf := bytes.Buffer{}
	for k, v := range *s {
		fmt.Fprintf(&buf, "%03d: %q\n", k, v)
	}
	return buf.String()
}
