package compiler

import (
  "fmt"
  "github.com/lestrrat/go-xslate/vm"
)

// Optimizer is the interface of things that can optimize the ByteCode
type Optimizer interface {
  Optimize(*vm.ByteCode) error
}

// NaiveOptimizer is the default ByteCode optimizer
type NaiveOptimizer struct {}

// Optimize modifies the ByteCode in place to an optimized version
func (o *NaiveOptimizer) Optimize(bc *vm.ByteCode) error {
  for i := 0; i < bc.Len(); i++ {
    op := bc.Get(i)
    if op == nil {
      return fmt.Errorf("failed to fetch op %d", i)
    }
    switch op.Type() {
    case vm.TXOPLiteral:
      if i + 1 < bc.Len() && bc.Get(i + 1).Type() == vm.TXOPPrintRaw {
        bc.OpList[i] = vm.NewOp(vm.TXOPPrintRawConst, op.ArgString())
        bc.OpList[i + 1] = vm.NewOp(vm.TXOPNoop)
        i++
      }
    }
  }
  return nil
}