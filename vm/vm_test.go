package vm

import (
  "testing"
)

func TestBasic(t *testing.T) {
  vm := NewVM()

  vm.st.Vars().Set("name", "Bob")
  pc := vm.st.pc
  pc.Append(&Op { TXCODE_noop, nil })
  pc.Append(&Op { TXCODE_literal, "Hello, World! " })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_fetch_s, "name" })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  // for debug only
  t.Logf("%s", vm.st.pc)

  vm.Run()

  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }

  expected := "Hello, World! Bob"
  if output != expected {
    t.Errorf("Expected output '%s', got '%s'", expected, output)
  }
}