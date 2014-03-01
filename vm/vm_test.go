package vm

import (
  "bytes"
  "testing"
)

func assertOutput(t *testing.T, vm *VM, expected string) {
  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }

  if output != expected {
    t.Errorf("Expected output '%s', got '%s'", expected, output)
  }
}

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

  assertOutput(t, vm, "Hello, World! Bob")
}

func TestFetchField(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  vm.st.Vars().Set("foo", struct { Value int } { 100 })
  pc.Append(&Op { TXCODE_fetch_s, "foo" })
  pc.Append(&Op { TXCODE_fetch_field_s, "value" })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  vm.Run()

  assertOutput(t, vm, "100")
}

func TestNonExistingSymbol(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.Append(&Op { TXCODE_fetch_s, "foo" })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  buf := &bytes.Buffer {}
  vm.st.warn = buf

  vm.Run()

  expected := "Use of nil to print\n"
  if warnOutput := buf.String(); warnOutput != expected {
    t.Errorf("Expected warning to be '%s', got '%s'", expected, warnOutput)
  }
}

func TestVm_Lvar(t *testing.T) {
  vm := NewVM()
  vm.st.PushFrame(NewFrame())
  pc := vm.st.pc
  pc.Append(&Op { TXCODE_literal, 999 })
  pc.Append(&Op { TXCODE_save_to_lvar, 0 })
  pc.Append(&Op { TXCODE_load_lvar, 0 })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  vm.Run()

  assertOutput(t, vm, "999")
}

func TestVM_Add(t *testing.T) {
  vm := NewVM()
  vm.st.PushFrame(NewFrame())
  pc := vm.st.pc
  pc.Append(&Op { TXCODE_literal, 999 })
  pc.Append(&Op { TXCODE_move_to_sb, nil })
  pc.Append(&Op { TXCODE_literal, 1 })
  pc.Append(&Op { TXCODE_add, nil })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  vm.Run()

  assertOutput(t, vm, "1000")
}

func TestVM_Sub(t *testing.T) {
  vm := NewVM()
  vm.st.PushFrame(NewFrame())
  pc := vm.st.pc
  pc.Append(&Op { TXCODE_literal, 999 })
  pc.Append(&Op { TXCODE_move_to_sb, nil })
  pc.Append(&Op { TXCODE_literal, 1 })
  pc.Append(&Op { TXCODE_sub, nil })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  vm.Run()

  assertOutput(t, vm, "998")
}



