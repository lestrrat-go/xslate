package vm

import (
  "reflect"
  "testing"
)

func TestVars_Basic(t *testing.T) {
  v := Vars {}
  v.Set("foo", 1)
  x, _ := v.Get("foo")
  if reflect.TypeOf(x).Kind() != reflect.Int {
    t.Errorf("Expected Get to return int, got %s", reflect.TypeOf(x).Kind())
  }
}