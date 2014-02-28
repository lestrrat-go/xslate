package vm

import (
  "testing"
)

func TestBasic(t *testing.T) {
  vm := NewVM()

  vm.st.pc = append(vm.st.pc, &Op { TXCODE_noop, nil })
  vm.st.pc = append(vm.st.pc, &Op { TXCODE_literal, "Hello, World" })
  vm.st.pc = append(vm.st.pc, &Op { TXCODE_print_raw, nil })
  vm.st.pc = append(vm.st.pc, &Op { TXCODE_end, nil })
  vm.Run()

  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }
  t.Logf(output)
}