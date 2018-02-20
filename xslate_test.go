package xslate

import (
	"fmt"
	"github.com/lestrrat-go/xslate/test"
	"log"
	"os"
	"reflect"
	"regexp"
	"testing"
)

type testctx struct {
	*test.Ctx
	XslateArgs Args
}

func newTestCtx(t test.Tester) *testctx {
	c := &testctx{
		test.NewCtx(t),
		nil,
	}

	// generate Xslate arguments
	args := Args{
		"Parser": Args{
			"Syntax": "TTerse",
		},
		"Loader": Args{
			"LoadPaths":  []string{c.BaseDir},
			"CacheDir":   c.Mkpath("cache"),
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

func ExampleXslate() {
	tx, err := New(Args{
		"Parser": Args{
			"Syntax": "TTerse",
		},
		"Loader": Args{
			"LoadPaths": []string{"/path/to/templates"},
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

func (c *testctx) renderString(template string, vars Vars) (string, error) {
	x := c.CreateTx()
	c.Logf("Rendering template '%s'", template)
	return x.RenderString(template, vars)
}

func (c *testctx) renderStringAndCompare(template string, vars Vars, expected interface{}) {
	c.Logf("Using template '%s', with vars '%#v'", template, vars)
	output, err := c.renderString(template, vars)

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

func (c *testctx) compareTemplateOutput(output string, expected interface{}) {
	typ := reflect.TypeOf(expected)
	switch typ.Kind() {
	case reflect.String:
		if output != expected.(string) {
			c.Errorf("Expected '%s', got '%s'", expected, output)
		}
		return
	case reflect.Ptr:
		if typ.Elem().Name() == "Regexp" {
			if !expected.(*regexp.Regexp).MatchString(output) {
				c.Errorf("Expected '%s', got '%s'", expected, output)
			}
			return
		}
	}
	c.Errorf("Unknown 'expected' type: %s", typ.Name())
}

func TestXslate_New_ParserSyntax(t *testing.T) {
	var err error
	_, err = New(Args{"Parser": Args{"Syntax": "Kolonish"}})
	if err != nil {
		t.Errorf("Expected Syntax: Kolonish to succeed, but got err: %s", err)
	}

	_, err = New(Args{"Parser": Args{"Syntax": "TTerse"}})
	if err != nil {
		t.Errorf("Expected Syntax: TTerse to succeed, but got err: %s", err)
	}
}
