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
  pc.AppendNoop()
  pc.AppendLiteral("Hello, World! ")
  pc.AppendPrintRaw()
  pc.AppendFetchSymbol("name")
  pc.AppendPrintRaw()
  pc.AppendEnd()

  // for debug only
  t.Logf("%s", vm.st.pc)
  // v, _ := json.Marshal(vm.st.pc)
  // t.Logf("json: %s", v)

  vm.Run()

  assertOutput(t, vm, "Hello, World! Bob")
}

func TestFetchField(t *testing.T) {
  vm := NewVM()
  vm.st.Vars().Set("foo", struct { Value int } { 100 })
  pc := vm.st.pc
  pc.AppendFetchSymbol("foo")
  pc.AppendFetchField("value")
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "100")
}

func TestNonExistingSymbol(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendFetchSymbol("foo")
  pc.AppendPrintRaw()
  pc.AppendEnd()

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
  pc := vm.st.pc
  pc.AppendLiteral(999)
  pc.AppendSaveToLvar(0)
  pc.AppendLoadLvar(0)
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "999")
}

func TestVM_AddInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(999)
  pc.AppendMoveToSb()
  pc.AppendLiteral(1)
  pc.AppendAdd()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "1000")
}

func TestVM_AddFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(0.999)
  pc.AppendMoveToSb()
  pc.AppendLiteral(0.001)
  pc.AppendAdd()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "1")
}

func TestVM_SubInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(999)
  pc.AppendMoveToSb()
  pc.AppendLiteral(1)
  pc.AppendSub()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "998")
}

func TestVM_SubFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(1)
  pc.AppendMoveToSb()
  pc.AppendLiteral(0.1)
  pc.AppendSub()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "0.9")
}

func TestVM_MulInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(3)
  pc.AppendMoveToSb()
  pc.AppendLiteral(4)
  pc.AppendMul()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "12")
}

func TestVM_MulFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(2.2)
  pc.AppendMoveToSb()
  pc.AppendLiteral(4)
  pc.AppendMul()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "8.8")
}

func TestVM_DivInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(6)
  pc.AppendMoveToSb()
  pc.AppendLiteral(3)
  pc.AppendDiv()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "2")
}

func TestVM_DivFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(10)
  pc.AppendMoveToSb()
  pc.AppendLiteral(4)
  pc.AppendDiv()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "2.5")
}

func TestVM_LvarAssignArithmeticResult(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(1)
  pc.AppendSaveToLvar(0)
  pc.AppendLiteral(2)
  pc.AppendSaveToLvar(1)
  pc.AppendLoadLvar(0)
  pc.AppendMoveToSb()
  pc.AppendLoadLvar(1)
  pc.AppendAdd()
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "3")
}

func TestVM_IfNoElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(true)
  pc.AppendAnd(3)
  pc.AppendLiteral("Hello, World!")
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).u_arg = false

  vm.Run()
  assertOutput(t, vm, "")
}

func TestVM_IfElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendLiteral(true)
  pc.AppendAnd(4)
  pc.AppendLiteral("Hello, World!")
  pc.AppendPrintRaw()
  pc.AppendGoto(3)
  pc.AppendLiteral("Ola, Mundo!")
  pc.AppendPrintRaw()
  pc.AppendEnd()

  vm.Run()

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).u_arg = false

  vm.Run()
  assertOutput(t, vm, "Ola, Mundo!")
}

