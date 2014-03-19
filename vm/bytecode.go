package vm

import (
  "bytes"
  "fmt"
  "time"
)

type ByteCode struct {
  GeneratedOn time.Time
  OpList      []*Op
}

func NewByteCode() *ByteCode {
  return &ByteCode { time.Now(), []*Op {} }
}

func (b *ByteCode) Len() int {
  return len(b.OpList)
}

func (b *ByteCode) Get(i int) *Op {
  return b.OpList[i]
}

func (b *ByteCode) Append(op *Op) {
  b.OpList = append(b.OpList, op)
}

func (b *ByteCode) String() string {
  buf := &bytes.Buffer {}

  fmt.Fprintf(buf, "Bytecode Generated On: %s\n", b.GeneratedOn)
  for k, v := range b.OpList {
    fmt.Fprintf(buf, "%03d. %s\n", k + 1, v)
  }
  return buf.String()
}

func (b *ByteCode) AppendOp(o OpType, args ...interface{}) *Op {
  x := NewOp(o, args...)
  b.Append(x)
  return x
}

