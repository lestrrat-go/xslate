package parser

import (
	"github.com/lestrrat/go-lex"
	"strings"
	"testing"
)

func TestItem_String(t *testing.T) {
	for i := lex.ItemDefaultMax + 1; i < DefaultItemTypeMax; i++ {
		it := lex.ItemType(i)
		if strings.HasPrefix(it.String(), "Unknown") {
			t.Errorf("%#v does not have String() implemented", it)
		}
	}
}
