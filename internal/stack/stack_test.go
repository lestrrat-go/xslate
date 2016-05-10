package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type IntWrap struct{ i int }

func TestStack_Grow(t *testing.T) {
	s := New(5)
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			s.Push(i)
		} else {
			s.Push(IntWrap{i})
		}
	}

	for i := 0; i < 10; i++ {
		x, err := s.Get(i)
		if !assert.NoError(t, err, "s.Get(%d) should succeed", i) {
			return
		}

		if i%2 == 0 {
			if !assert.Equal(t, i, x, "s.Get(%d) should be %d (got %v)", i, i, x) {
				t.Logf("%s", s)
				return
			}
		} else {
			if !assert.Equal(t, IntWrap{i}, x, "s.Get(%d) should be IntWrap{%d} (got %v)", i, i, x) {
				t.Logf("%s", s)
				return
			}
		}
	}

	for i := 9; i > -1; i-- {
		x := s.Pop()
		if i%2 == 0 {
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
