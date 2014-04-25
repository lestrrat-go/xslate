package vm

import (
	"fmt"
	"testing"
	"time"
)

func TestByteCode_Len(t *testing.T) {
	bc := &ByteCode{}
	if bc.Len() != 0 {
		t.Errorf("Expected len == 0, got %d\n", bc.Len())
	}

	bc.AppendOp(TXOPEnd)
	if bc.Len() != 1 {
		t.Errorf("Expected len == 1, got %d\n", bc.Len())
	}
}

func TestByteCode_String(t *testing.T) {
	now := time.Now()
	bc := &ByteCode{Name: "DUMMY", GeneratedOn: now}
	expected := fmt.Sprintf("// Bytecode for 'DUMMY'\n// Generated On: %s\n", now)
	if bc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bc.String())
	}

	bc.AppendOp(TXOPEnd)
	expected = fmt.Sprintf("// Bytecode for 'DUMMY'\n// Generated On: %s\n001. Op[end]\n", now)
	if bc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bc.String())
	}
}
