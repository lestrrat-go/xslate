package vm

import (
	"fmt"
	"time"

	"github.com/lestrrat/go-xslate/internal/rbpool"
)

// NewByteCode creates an empty ByteCode instance.
func NewByteCode() *ByteCode {
	return &ByteCode{
		GeneratedOn: time.Now(),
		Name:        "",
		OpList:      nil,
		Version:     1.0,
	}
}

// Len returns the number of op codes in this ByteCode instance
func (b *ByteCode) Len() int {
	return len(b.OpList)
}

// Get returns an vm.Op struct at location i. No check is performed to see
// if this index is valid
func (b *ByteCode) Get(i int) Op {
	return b.OpList[i]
}

// Append appends an op code to the current list of op codes.
func (b *ByteCode) Append(op Op) {
	b.OpList = append(b.OpList, op)
}

// AppendOp is an utility method to create AND append a new op code to the
// current list of op codes
func (b *ByteCode) AppendOp(o OpType, args ...interface{}) Op {
	x := NewOp(o, args...)
	b.Append(x)
	return x
}

// String returns the textual representation of this ByteCode
func (b *ByteCode) String() string {
	buf := rbpool.Get()
	defer rbpool.Release(buf)

	fmt.Fprintf(buf,
		"// Bytecode for '%s'\n// Generated On: %s\n",
		b.Name,
		b.GeneratedOn,
	)
	for k, v := range b.OpList {
		fmt.Fprintf(buf, "%03d. %s\n", k+1, v)
	}
	return buf.String()
}
