package vm

import (
  "bytes"
  "fmt"
  "time"
)

// ByteCode is the collection of op codes that the Xslate Virtual Machine
// should run. It is created from a compiler.Compiler
type ByteCode struct {
  GeneratedOn time.Time
  OpList      []*Op
}

// NewByteCode creates an empty ByteCode instance.
func NewByteCode() *ByteCode {
  return &ByteCode { time.Now(), []*Op {} }
}

// Len returns the number of op codes in this ByteCode instance
func (b *ByteCode) Len() int {
  return len(b.OpList)
}

// Get returns an vm.Op struct at location i. No check is performed to see
// if this index is valid
func (b *ByteCode) Get(i int) *Op {
  return b.OpList[i]
}

// Append appends an op code to the current list of op codes.
func (b *ByteCode) Append(op *Op) {
  b.OpList = append(b.OpList, op)
}

// AppendOp is an utility method to create AND append a new op code to the
// current list of op codes
func (b *ByteCode) AppendOp(o OpType, args ...interface{}) *Op {
  x := NewOp(o, args...)
  b.Append(x)
  return x
}

// String returns the textual representation of this ByteCode
func (b *ByteCode) String() string {
  buf := &bytes.Buffer {}

  fmt.Fprintf(buf, "Bytecode Generated On: %s\n", b.GeneratedOn)
  for k, v := range b.OpList {
    fmt.Fprintf(buf, "%03d. %s\n", k + 1, v)
  }
  return buf.String()
}

