package vm

import (
  "bytes"
  "fmt"
  "reflect"
  "regexp"
  "testing"
  "time"
)

func assertOutput(t *testing.T, vm *VM, expected interface {}) {
  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }

  vtype := reflect.TypeOf(expected)
  switch {
  case vtype.Kind() == reflect.String:
    if output != expected.(string) {
      t.Errorf("Expected output '%s', got '%s'", expected, output)
    }
  case vtype.Kind() == reflect.Ptr && vtype.Elem().Kind() == reflect.Struct && vtype.Elem().Name() == "Regexp":
    if ! expected.(*regexp.Regexp).MatchString(output) {
      t.Errorf("Expected output to match '%s', got '%s'", expected, output)
    }
  default:
    panic(fmt.Sprintf("Can't handle type %s", vtype.Kind()))
  }
}

func TestBasic(t *testing.T) {
  vm := NewVM()

  vm.st.Vars().Set("name", "Bob")
  pc := vm.st.pc
  pc.AppendOp(TXOP_noop)
  pc.AppendOp(TXOP_literal, "Hello, World! ")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_fetch_s, "name")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "Hello, World! Bob")
}

func TestFetchField(t *testing.T) {
  vm := NewVM()
  vm.st.Vars().Set("foo", struct { Value int } { 100 })
  pc := vm.st.pc
  pc.AppendOp(TXOP_fetch_s, "foo")
  pc.AppendOp(TXOP_fetch_field_s, "value")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "100")
}

func TestNonExistingSymbol(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_fetch_s, "foo")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

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
  pc.AppendOp(TXOP_literal, 999)
  pc.AppendOp(TXOP_save_to_lvar, 0)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "999")
}

func TestVM_AddInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 999)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_add)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "1000")
}

func TestVM_AddFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 0.999)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 0.001)
  pc.AppendOp(TXOP_add)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "1")
}

func TestVM_SubInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 999)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_sub)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "998")
}

func TestVM_SubFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 0.1)
  pc.AppendOp(TXOP_sub)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "0.9")
}

func TestVM_MulInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 3)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 4)
  pc.AppendOp(TXOP_mul)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "12")
}

func TestVM_MulFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 2.2)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 4)
  pc.AppendOp(TXOP_mul)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "8.8")
}

func TestVM_DivInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 6)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 3)
  pc.AppendOp(TXOP_div)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "2")
}

func TestVM_DivFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 10)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 4)
  pc.AppendOp(TXOP_div)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "2.5")
}

func TestVM_LvarAssignArithmeticResult(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_save_to_lvar, 0)
  pc.AppendOp(TXOP_literal, 2)
  pc.AppendOp(TXOP_save_to_lvar, 1)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_load_lvar, 1)
  pc.AppendOp(TXOP_add)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "3")
}

func TestVM_IfNoElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, true)
  pc.AppendOp(TXOP_and, 3)
  pc.AppendOp(TXOP_literal, "Hello, World!")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).u_arg = false

  vm.Run()
  assertOutput(t, vm, "")
}

func TestVM_IfElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, true)
  pc.AppendOp(TXOP_and, 4)
  pc.AppendOp(TXOP_literal, "Hello, World!")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_goto, 3)
  pc.AppendOp(TXOP_literal, "Ola, Mundo!")
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).u_arg = false

  vm.Run()
  assertOutput(t, vm, "Ola, Mundo!")
}

func TestVM_ForLoop(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, []int { 10, 9, 8, 7, 6, 5, 4, 3, 2, 1 })
  pc.AppendOp(TXOP_for_start, 0)
  pc.AppendOp(TXOP_literal, 0)
  pc.AppendOp(TXOP_for_iter, 4)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_goto, -4)
  pc.AppendOp(TXOP_end)

  vm.Run()

  assertOutput(t, vm, "10987654321")
}

func TestVM_HtmlEscape(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, "<div>Hello, World!</div>")
  pc.AppendOp(TXOP_html_escape)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, "&lt;div&gt;Hello, World!&lt;/div&gt;")
}

func TestVM_UriEscape(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, "日本語")
  pc.AppendOp(TXOP_uri_escape)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, "%E6%97%A5%E6%9C%AC%E8%AA%9E")
}

func TestVM_Eq(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_eq)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)
  vm.Run()
  assertOutput(t, vm, "true")

  pc.Get(0).u_arg = 2
  vm.Run()
  assertOutput(t, vm, "false")
}

func TestVM_Ne(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_move_to_sb)
  pc.AppendOp(TXOP_literal, 1)
  pc.AppendOp(TXOP_ne)
  pc.AppendOp(TXOP_print_raw)
  pc.AppendOp(TXOP_end)
  vm.Run()
  assertOutput(t, vm, "false")

  pc.Get(0).u_arg = 2
  vm.Run()
  assertOutput(t, vm, "true")
}

func TestVM_MarkRaw(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOP_literal, "<div>Hello</div>")
  pc.AppendOp(TXOP_mark_raw)
  pc.AppendOp(TXOP_print)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, "<div>Hello</div>")
}

func TestVM_FunCall(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  pc.AppendOp(TXOP_literal, time.Now)
  pc.AppendOp(TXOP_save_to_lvar, 0)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_funcall, nil)
  pc.AppendOp(TXOP_print)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_FunCallFromDepot(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  fd := FuncDepot {}
  fd.Set("Now", time.Now)
  pc.AppendOp(TXOP_literal, fd)
  pc.AppendOp(TXOP_save_to_lvar, 0)
  pc.AppendOp(TXOP_pushmark)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_push)
  pc.AppendOp(TXOP_funcall, "Now")
  pc.AppendOp(TXOP_popmark)
  pc.AppendOp(TXOP_print)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_MethodCall(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  // tx.Render(..., &Vars { t: time.Now() })
  // [% t.Before(time.Now()) %]
  pc.AppendOp(TXOP_literal, time.Now())
  pc.AppendOp(TXOP_save_to_lvar, 0)
  pc.AppendOp(TXOP_pushmark)
  pc.AppendOp(TXOP_load_lvar, 0)
  pc.AppendOp(TXOP_push)
  pc.AppendOp(TXOP_literal, time.Now())
  pc.AppendOp(TXOP_push)
  pc.AppendOp(TXOP_methodcall, "Before")
  pc.AppendOp(TXOP_popmark)
  pc.AppendOp(TXOP_print)
  pc.AppendOp(TXOP_end)

  vm.Run()
  assertOutput(t, vm, "false")
}

