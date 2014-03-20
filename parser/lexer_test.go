package parser

import (
  "strings"
  "testing"
)

func TestItem_String(t *testing.T) {
  for i := 0; i < int(DefaultItemTypeMax); i++ {
    it := LexItemType(i)
    if strings.HasPrefix(it.String(), "Unknown") {
      t.Errorf("%#v does not have String() implemented", it)
    }
  }
}
