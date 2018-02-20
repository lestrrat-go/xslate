package frame

import (
	"testing"

	"github.com/lestrrat-go/xslate/internal/stack"
	"github.com/stretchr/testify/assert"
)

func TestFrame_Lvar(t *testing.T) {
	f := New(stack.New(5))
	f.SetLvar(0, 1)
	x, err := f.GetLvar(0)
	if !assert.NoError(t, err, "f.GetLvar(0) should succeed") {
		return
	}

	if !assert.Equal(t, 1, x, "f.GetLvar(0) should be 1") {
		return
	}
}
