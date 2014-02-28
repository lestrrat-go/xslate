package vm

import (
  "bytes"
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

func TestFetchField(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  vm.st.Vars().Set("foo", struct { Value int } { 100 })
  pc.Append(&Op { TXCODE_fetch_s, "foo" })
  pc.Append(&Op { TXCODE_fetch_field_s, "Value" })
  pc.Append(&Op { TXCODE_print_raw, nil })
  pc.Append(&Op { TXCODE_end, nil })

  vm.Run()

  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }

  expected := "100"
  if output != expected {
    t.Errorf("Expected output '%s', got '%s'", expected, output)
  }
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