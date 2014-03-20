package parser

import (
  "strings"
  "testing"
)

func TestNode_String(t *testing.T) {
  for i := 0; i < int(NodeMax); i++ {
    nt := NodeType(i)
    if strings.HasPrefix(nt.String(), "Unknown") {
      t.Errorf("%#v does not have String() implemented", nt)
    }
  }
}
