package vm

import (
  "bytes"
  "fmt"
)

type ByteCode []*Op

func (l *ByteCode) Len() int {
  return len(*l)
}

func (l *ByteCode) Get(i int) *Op {
  return (*l)[i]
}

func (l *ByteCode) Append(op *Op) {
  *l = append(*l, op)
}

func (l *ByteCode) String() string {
  buf := &bytes.Buffer {}
  for k, v := range *l {
    buf.WriteString(fmt.Sprintf("%03d. %s\n", k + 1, v))
  }
  return buf.String()
}

func (l *ByteCode) AppendOp(o OpType, args ...interface{}) *Op {
  x := NewOp(o, args...)
  l.Append(x)
  return x
}

