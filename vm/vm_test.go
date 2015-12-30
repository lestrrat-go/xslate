package vm

import (
	"bytes"
	"fmt"
	txtime "github.com/lestrrat/go-xslate/functions/time"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func assertOutput(t *testing.T, bc *ByteCode, vars Vars, expected interface{}) {
	buf := &bytes.Buffer{}
	vm := NewVM()
	vm.Run(bc, vars, buf)
	output := buf.String()

	vtype := reflect.TypeOf(expected)
	switch {
	case vtype.Kind() == reflect.String:
		if output != expected.(string) {
			t.Errorf("Expected output '%s', got '%s'", expected, output)
		}
	case vtype.Kind() == reflect.Ptr && vtype.Elem().Kind() == reflect.Struct && vtype.Elem().Name() == "Regexp":
		if !expected.(*regexp.Regexp).MatchString(output) {
			t.Errorf("Expected output to match '%s', got '%s'", expected, output)
		}
	default:
		panic(fmt.Sprintf("Can't handle type %s", vtype.Kind()))
	}
}

func TestBasic(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPNoop)
	bc.AppendOp(TXOPLiteral, "Hello, World! ")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPFetchSymbol, "name")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, Vars{"name": "Bob"}, "Hello, World! Bob")
}

func TestFetchField(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPFetchSymbol, "foo")
	bc.AppendOp(TXOPFetchFieldSymbol, "value")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, Vars{"foo": struct{ Value int }{100}}, "100")
}

func TestNonExistingSymbol(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPFetchSymbol, "foo")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	buf := &bytes.Buffer{}
	vm := NewVM()
	vm.st.warn = buf

	vm.Run(bc, nil, &bytes.Buffer{})

	expected := "Use of nil to print\n"
	if warnOutput := buf.String(); warnOutput != expected {
		t.Errorf("Expected warning to be '%s', got '%s'", expected, warnOutput)
	}
}

func TestVm_Lvar(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 999)
	bc.AppendOp(TXOPSaveToLvar, 0)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "999")
}

func TestVM_AddInt(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 999)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPAdd)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "1000")
}

func TestVM_AddFloat(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 0.999)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 0.001)
	bc.AppendOp(TXOPAdd)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "1")
}

func TestVM_SubInt(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 999)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPSub)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "998")
}

func TestVM_SubFloat(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 0.1)
	bc.AppendOp(TXOPSub)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "0.9")
}

func TestVM_MulInt(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 3)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 4)
	bc.AppendOp(TXOPMul)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "12")
}

func TestVM_MulFloat(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 2.2)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 4)
	bc.AppendOp(TXOPMul)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "8.8")
}

func TestVM_DivInt(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 6)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 3)
	bc.AppendOp(TXOPDiv)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "2")
}

func TestVM_DivFloat(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 10)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 4)
	bc.AppendOp(TXOPDiv)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "2.5")
}

func TestVM_LvarAssignArithmeticResult(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPSaveToLvar, 0)
	bc.AppendOp(TXOPLiteral, 2)
	bc.AppendOp(TXOPSaveToLvar, 1)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLoadLvar, 1)
	bc.AppendOp(TXOPAdd)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "3")
}

func TestVM_IfNoElse(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, true)
	bc.AppendOp(TXOPAnd, 3)
	bc.AppendOp(TXOPLiteral, "Hello, World!")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "Hello, World!")

	bc.Get(0).SetArg(false)

	assertOutput(t, bc, nil, "")
}

func TestVM_IfElse(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, true)
	bc.AppendOp(TXOPAnd, 4)
	bc.AppendOp(TXOPLiteral, "Hello, World!")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPGoto, 3)
	bc.AppendOp(TXOPLiteral, "Ola, Mundo!")
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "Hello, World!")

	bc.Get(0).SetArg(false)

	assertOutput(t, bc, nil, "Ola, Mundo!")
}

func TestVM_ForLoop(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	bc.AppendOp(TXOPForStart, 0)
	bc.AppendOp(TXOPLiteral, 0)
	bc.AppendOp(TXOPForIter, 4)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPGoto, -4)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "10987654321")
}

func TestVM_HtmlEscape(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, "<div>Hello, World!</div>")
	bc.AppendOp(TXOPHTMLEscape)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "&lt;div&gt;Hello, World!&lt;/div&gt;")
}

func TestVM_UriEscape(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, "日本語")
	bc.AppendOp(TXOPUriEscape)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "%E6%97%A5%E6%9C%AC%E8%AA%9E")
}

func TestVM_Eq(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPEquals)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)
	assertOutput(t, bc, nil, "true")

	bc.Get(0).SetArg(2)
	assertOutput(t, bc, nil, "false")
}

func TestVM_Ne(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPNotEquals)
	bc.AppendOp(TXOPPrintRaw)
	bc.AppendOp(TXOPEnd)
	assertOutput(t, bc, nil, "false")

	bc.Get(0).SetArg(2)
	assertOutput(t, bc, nil, "true")
}

func TestVM_MarkRaw(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, "<div>Hello</div>")
	bc.AppendOp(TXOPMarkRaw)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "<div>Hello</div>")
}

func TestVM_FunCall(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPLiteral, time.Now)
	bc.AppendOp(TXOPSaveToLvar, 0)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPFunCall, nil)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_FunCallFromDepot(t *testing.T) {
	bc := NewByteCode()
	fd := txtime.Depot()
	bc.AppendOp(TXOPLiteral, fd)
	bc.AppendOp(TXOPSaveToLvar, 0)
	bc.AppendOp(TXOPPushmark)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPFunCallSymbol, "Now")
	bc.AppendOp(TXOPPopmark)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ [+-]\d{4} \w+`))
}

func TestVM_MethodCall(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Nanosecond)

	bc := NewByteCode()
	// tx.Render(..., &Vars { t: t1 })
	// [% t.Before(t2) %]
	bc.AppendOp(TXOPLiteral, t1)
	bc.AppendOp(TXOPSaveToLvar, 0)
	bc.AppendOp(TXOPPushmark)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPLiteral, t2)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPMethodCall, "Before")
	bc.AppendOp(TXOPPopmark)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "true")
}

func TestVM_RangeMakeArray(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPPushmark)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPMoveToSb)
	bc.AppendOp(TXOPLiteral, 10)
	bc.AppendOp(TXOPRange)
	bc.AppendOp(TXOPMakeArray)
	bc.AppendOp(TXOPForStart, 0)
	bc.AppendOp(TXOPLiteral, 0)
	bc.AppendOp(TXOPForIter, 6)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPLiteral, ",")
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPGoto, -6)
	bc.AppendOp(TXOPEnd)

	assertOutput(t, bc, nil, "1,2,3,4,5,6,7,8,9,10,")
}

func TestVM_MakeHash(t *testing.T) {
	bc := NewByteCode()
	bc.AppendOp(TXOPPushmark)
	bc.AppendOp(TXOPLiteral, "foo")
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPLiteral, 1)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPLiteral, "bar")
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPLiteral, 2)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPMakeHash)
	bc.AppendOp(TXOPPopmark)
	bc.AppendOp(TXOPPushmark)
	bc.AppendOp(TXOPPush)
	bc.AppendOp(TXOPMethodCall, "Keys")
	bc.AppendOp(TXOPPopmark)
	bc.AppendOp(TXOPForStart, 0)
	bc.AppendOp(TXOPLiteral, 0)
	bc.AppendOp(TXOPForIter, 6)
	bc.AppendOp(TXOPLoadLvar, 0)
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPLiteral, ",")
	bc.AppendOp(TXOPPrint)
	bc.AppendOp(TXOPGoto, -6)
	bc.AppendOp(TXOPEnd)

	buf := &bytes.Buffer{}
	vm := NewVM()
	vm.Run(bc, nil, buf)
	// Note: order of keys may change depending on environment..., so we can't
	// just compare against vm.Output()
	output := buf.String()

	for _, v := range []string{"foo,", "bar,"} {
		if !strings.Contains(output, v) {
			t.Errorf("Expected to find '%s', but did not find it in '%s'", v, output)
		}
	}
}
