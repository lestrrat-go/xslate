package xslate

import (
  "github.com/lestrrat/go-xslate/test"
  "testing"
)

func newKolonCtx(t test.Tester) *testctx {
  c := newTestCtx(t)
  pargs := c.XslateArgs["Parser"].(Args)
  pargs["Syntax"] = "Kolon"

  return c
}

func TestKolonish_SimpleString(t *testing.T) {
  c := newKolonCtx(t)
  defer c.Cleanup()

  c.renderStringAndCompare(`Hello, World!`, nil, `Hello, World!`)
  c.renderStringAndCompare(`    <:- "Hello, World!" :>`, nil, `Hello, World!`)
  c.renderStringAndCompare(`<: "Hello, World!" -:>    `, nil, `Hello, World!`)
}