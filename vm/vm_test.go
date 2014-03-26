package vm

import (
  "bytes"
  "fmt"
  "reflect"
  "regexp"
  "strings"
  "testing"
  "time"
  txtime "github.com/lestrrat/go-xslate/functions/time"
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
  pc.AppendOp(TXOPNoop)
  pc.AppendOp(TXOPLiteral, "Hello, World! ")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPFetchSymbol, "name")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "Hello, World! Bob")
}

func TestFetchField(t *testing.T) {
  vm := NewVM()
  vm.st.Vars().Set("foo", struct { Value int } { 100 })
  pc := vm.st.pc
  pc.AppendOp(TXOPFetchSymbol, "foo")
  pc.AppendOp(TXOPFetchFieldSymbol, "value")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "100")
}

func TestNonExistingSymbol(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPFetchSymbol, "foo")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  buf := &bytes.Buffer {}
  vm.st.warn = buf

  vm.Run(nil, nil)

  expected := "Use of nil to print\n"
  if warnOutput := buf.String(); warnOutput != expected {
    t.Errorf("Expected warning to be '%s', got '%s'", expected, warnOutput)
  }
}

func TestVm_Lvar(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 999)
  pc.AppendOp(TXOPSaveToLvar, 0)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "999")
}

func TestVM_AddInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 999)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPAdd)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "1000")
}

func TestVM_AddFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 0.999)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 0.001)
  pc.AppendOp(TXOPAdd)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "1")
}

func TestVM_SubInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 999)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPSub)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "998")
}

func TestVM_SubFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 0.1)
  pc.AppendOp(TXOPSub)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "0.9")
}

func TestVM_MulInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 3)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 4)
  pc.AppendOp(TXOPMul)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "12")
}

func TestVM_MulFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 2.2)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 4)
  pc.AppendOp(TXOPMul)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "8.8")
}

func TestVM_DivInt(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 6)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 3)
  pc.AppendOp(TXOPDiv)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "2")
}

func TestVM_DivFloat(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 10)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 4)
  pc.AppendOp(TXOPDiv)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "2.5")
}

func TestVM_LvarAssignArithmeticResult(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPSaveToLvar, 0)
  pc.AppendOp(TXOPLiteral, 2)
  pc.AppendOp(TXOPSaveToLvar, 1)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLoadLvar, 1)
  pc.AppendOp(TXOPAdd)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "3")
}

func TestVM_IfNoElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, true)
  pc.AppendOp(TXOPAnd, 3)
  pc.AppendOp(TXOPLiteral, "Hello, World!")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).SetArg(false)

  vm.Run(nil, nil)
  assertOutput(t, vm, "")
}

func TestVM_IfElse(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, true)
  pc.AppendOp(TXOPAnd, 4)
  pc.AppendOp(TXOPLiteral, "Hello, World!")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPGoto, 3)
  pc.AppendOp(TXOPLiteral, "Ola, Mundo!")
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "Hello, World!")

  pc.Get(0).SetArg(false)

  vm.Run(nil, nil)
  assertOutput(t, vm, "Ola, Mundo!")
}

func TestVM_ForLoop(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, []int { 10, 9, 8, 7, 6, 5, 4, 3, 2, 1 })
  pc.AppendOp(TXOPForStart, 0)
  pc.AppendOp(TXOPLiteral, 0)
  pc.AppendOp(TXOPForIter, 4)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPGoto, -4)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)

  assertOutput(t, vm, "10987654321")
}

func TestVM_HtmlEscape(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, "<div>Hello, World!</div>")
  pc.AppendOp(TXOPHTMLEscape)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, "&lt;div&gt;Hello, World!&lt;/div&gt;")
}

func TestVM_UriEscape(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, "日本語")
  pc.AppendOp(TXOPUriEscape)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, "%E6%97%A5%E6%9C%AC%E8%AA%9E")
}

func TestVM_Eq(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPEquals)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)
  vm.Run(nil, nil)
  assertOutput(t, vm, "true")

  pc.Get(0).SetArg(2)
  vm.Run(nil, nil)
  assertOutput(t, vm, "false")
}

func TestVM_Ne(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPNotEquals)
  pc.AppendOp(TXOPPrintRaw)
  pc.AppendOp(TXOPEnd)
  vm.Run(nil, nil)
  assertOutput(t, vm, "false")

  pc.Get(0).SetArg(2)
  vm.Run(nil, nil)
  assertOutput(t, vm, "true")
}

func TestVM_MarkRaw(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc
  pc.AppendOp(TXOPLiteral, "<div>Hello</div>")
  pc.AppendOp(TXOPMarkRaw)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, "<div>Hello</div>")
}

func TestVM_FunCall(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  pc.AppendOp(TXOPLiteral, time.Now)
  pc.AppendOp(TXOPSaveToLvar, 0)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPFunCall, nil)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_FunCallFromDepot(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  fd := txtime.Depot()
  pc.AppendOp(TXOPLiteral, fd)
  pc.AppendOp(TXOPSaveToLvar, 0)
  pc.AppendOp(TXOPPushmark)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPFunCallSymbol, "Now")
  pc.AppendOp(TXOPPopmark)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_MethodCall(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  // tx.Render(..., &Vars { t: time.Now() })
  // [% t.Before(time.Now()) %]
  pc.AppendOp(TXOPLiteral, time.Now())
  pc.AppendOp(TXOPSaveToLvar, 0)
  pc.AppendOp(TXOPPushmark)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPLiteral, time.Now())
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPMethodCall, "Before")
  pc.AppendOp(TXOPPopmark)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, "true")
}

func TestVM_RangeMakeArray(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  pc.AppendOp(TXOPPushmark)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPMoveToSb)
  pc.AppendOp(TXOPLiteral, 10)
  pc.AppendOp(TXOPRange)
  pc.AppendOp(TXOPMakeArray)
  pc.AppendOp(TXOPForStart, 0)
  pc.AppendOp(TXOPLiteral, 0)
  pc.AppendOp(TXOPForIter, 6)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPLiteral, ",")
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPGoto, -6)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  assertOutput(t, vm, "1,2,3,4,5,6,7,8,9,10,")
}

func TestVM_MakeHash(t *testing.T) {
  vm := NewVM()
  pc := vm.st.pc

  pc.AppendOp(TXOPPushmark)
  pc.AppendOp(TXOPLiteral, "foo")
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPLiteral, 1)
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPLiteral, "bar")
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPLiteral, 2)
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPMakeHash)
  pc.AppendOp(TXOPPopmark)
  pc.AppendOp(TXOPPushmark)
  pc.AppendOp(TXOPPush)
  pc.AppendOp(TXOPMethodCall, "Keys")
  pc.AppendOp(TXOPPopmark)
  pc.AppendOp(TXOPForStart, 0)
  pc.AppendOp(TXOPLiteral, 0)
  pc.AppendOp(TXOPForIter, 6)
  pc.AppendOp(TXOPLoadLvar, 0)
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPLiteral, ",")
  pc.AppendOp(TXOPPrint)
  pc.AppendOp(TXOPGoto, -6)
  pc.AppendOp(TXOPEnd)

  vm.Run(nil, nil)
  // Note: order of keys may change depending on environment..., so we can't
  // just compare against vm.Output()
  output, err := vm.OutputString()
  if err != nil {
    t.Errorf("Error getting output: %s", err)
  }

  for _, v := range []string { "foo,", "bar," } {
    if ! strings.Contains(output, v) {
      t.Errorf("Expected to find '%s', but did not find it in '%s'", v, output)
    }
  }
}
