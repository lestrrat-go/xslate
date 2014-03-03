package vm

import (
  "bytes"
  "fmt"
)

type OpList []*Op

func (l *OpList) Get(i int) *Op {
  return (*l)[i]
}

func (l *OpList) Append(op *Op) {
  *l = append(*l, op)
}

func (l *OpList) String() string {
  buf := &bytes.Buffer {}
  for k, v := range *l {
    buf.WriteString(fmt.Sprintf("%03d. %s\n", k + 1, v))
  }
  return buf.String()
}

func (l *OpList) AppendOp(o OpType, args ...interface{}) {
  l.Append(NewOp(o, args...))
}

