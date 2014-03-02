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

func (l *OpList) AppendNoop() {
  l.AppendOp(TXOP_noop)
}

func (l *OpList) AppendEnd() {
  l.AppendOp(TXOP_end)
}

func (l *OpList) AppendLiteral(v interface {}) {
  l.AppendOp(TXOP_literal, v)
}

func (l *OpList) AppendPrintRaw() {
  l.Append(NewOp(TXOP_print_raw))
}

func (l *OpList) AppendFetchSymbol(name string) {
  l.Append(NewOp(TXOP_fetch_s, name))
}

func (l *OpList) AppendFetchField(name string) {
  l.Append(NewOp(TXOP_fetch_field_s, name))
}

func (l *OpList) AppendSaveToLvar(ix int) {
  l.Append(NewOp(TXOP_save_to_lvar, ix))
}

func (l *OpList) AppendLoadLvar(ix int) {
  l.Append(NewOp(TXOP_load_lvar, ix))
}

func (l *OpList) AppendMoveToSb() {
  l.Append(NewOp(TXOP_move_to_sb))
}

func (l *OpList) AppendAdd() {
  l.AppendOp(TXOP_add)
}

func (l *OpList) AppendSub() {
  l.AppendOp(TXOP_sub)
}

func (l *OpList) AppendMul() {
  l.AppendOp(TXOP_mul)
}

func (l *OpList) AppendDiv() {
  l.AppendOp(TXOP_div)
}

func (l *OpList) AppendAnd(ix int) {
  l.AppendOp(TXOP_and, ix)
}

func (l *OpList) AppendGoto(ix int) {
  l.AppendOp(TXOP_goto, ix)
}