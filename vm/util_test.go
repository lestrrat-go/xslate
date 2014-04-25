package vm

import (
	"reflect"
	"testing"
)

func TestNumericNormlization(t *testing.T) {
	var v reflect.Value

	v = interfaceToNumeric(1)
	if v.Kind() != reflect.Int {
		t.Errorf("Expected Kind() to be Int, got %s", v.Kind())
	}

	v = interfaceToNumeric(1.0)
	if v.Kind() != reflect.Float64 {
		t.Errorf("Expected Kind() to be Float, got %s", v.Kind())
	}

	// Non-numbers should all be 0
	thingies := []interface{}{
		"Hello, World",
		struct{ foo string }{"foo"},
		&struct{ bar int }{0},
	}
	for _, x := range thingies {
		v = interfaceToNumeric(x)
		if v.Kind() != reflect.Int {
			t.Errorf("Expected Kind() to be Int, got %s", v.Kind())
		} else {
			if v.Int() != 0 {
				t.Errorf("Expected v.Int() to be 0, got %d", v.Int())
			}
		}
	}
}

func TestAlignTypesForArithmetic(t *testing.T) {
	leftV, rightV := alignTypesForArithmetic(1, 1.0)

	if leftV.Kind() != rightV.Kind() {
		t.Errorf("leftV.Kind (%s) != rightV.Kind (%s)", leftV.Kind(), rightV.Kind())
	}

	if leftV.Kind() != reflect.Float64 {
		t.Errorf("leftV should have been upgraded to Float64, but got %s", leftV.Kind())
	}
}
