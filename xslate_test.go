package xslate

import (
  "fmt"
  "log"
  "os"
  "reflect"
  "regexp"
  "testing"
  "github.com/lestrrat/go-xslate/test"
)

type testctx struct {
  *test.Ctx
  XslateArgs Args
}

func newTestCtx(t test.Tester) *testctx {
  c := &testctx {
    test.NewCtx(t),
    nil,
  }

  // generate Xslate arguments
  args := Args {
    "Parser": Args {
      "Syntax": "TTerse",
    },
    "Loader": Args {
      "LoadPaths": []string { c.BaseDir },
      "CacheDir": c.Mkpath("cache"),
      "CacheLevel": 1,
    },
  }
  c.XslateArgs = args

  return c
}

func (c *testctx) CreateTx() *Xslate {
  tx, err := New(c.XslateArgs)
  if err != nil {
    c.Fatalf("error: failed to create Xslate instance: %s", err)
  }
  return tx
}

func ExampleXslate () {
  tx, err := New(Args {
    "Parser": Args {
      "Syntax": "TTerse",
    },
    "Loader": Args {
      "LoadPaths": []string { "/path/to/templates" },
    },
  })
  if err != nil {
    log.Fatalf("Failed to create Xslate: %s", err)
  }
  output, err := tx.Render("foo.tx", nil)
  if err != nil {
    log.Fatalf("Failed to render template: %s", err)
  }
  fmt.Fprintf(os.Stdout, output)
}

func (c *testctx) renderStringAndCompare(template string, vars Vars, expected interface {}) {
  x := c.CreateTx()
  output, err := x.RenderString(template, vars)

  if err != nil {
    c.Fatalf("Failed to render template: %s", err)
  }
  c.compareTemplateOutput(output, expected)
}

func (c *testctx) renderAndCompare(tx *Xslate, key string, vars Vars, expected string) {
  output, err := tx.Render(key, vars)
  if err != nil {
    c.Fatalf("Failed to render template: %s", err)
  }
  c.compareTemplateOutput(output, expected)
}

func (c *testctx) compareTemplateOutput(output string, expected interface {}) {
  typ := reflect.TypeOf(expected)
  switch typ.Kind() {
  case reflect.String:
    if output != expected.(string) {
      c.Errorf("Expected '%s', got '%s'", expected, output)
    }
    return
  case reflect.Ptr:
    if typ.Elem().Name() == "Regexp" {
      if ! expected.(*regexp.Regexp).MatchString(output) {
        c.Errorf("Expected '%s', got '%s'", expected, output)
      }
      return
    }
  }
  c.Errorf("Unknown 'expected' type: %s", typ.Name())
}

func TestXslate_New_ParserSyntax(t *testing.T) {
  var err error
  _, err = New(Args { "Parser": Args { "Syntax": "Kolonish" } })
  if err != nil {
    t.Errorf("Expected Syntax: Kolonish to succeed, but got err: %s", err)
  }

  _, err = New(Args { "Parser": Args { "Syntax": "TTerse" } })
  if err != nil {
    t.Errorf("Expected Syntax: TTerse to succeed, but got err: %s", err)
  }
}

