package compiler

import (
	"github.com/lestrrat-go/xslate/parser/tterse"
	"github.com/lestrrat-go/xslate/test"
	"github.com/lestrrat-go/xslate/vm"
	"testing"
)

func compile(t *testing.T, tmpl string) *vm.ByteCode {
	p := tterse.New()
	ast, err := p.ParseString(tmpl, tmpl)
	if err != nil {
		t.Fatalf("Failed to parse template: %s", err)
	}

	c := New()
	bc, err := c.Compile(ast)
	if err != nil {
		t.Fatalf("Failed to compile ast: %s", err)
	}

	t.Logf("-> %+v", bc)

	return bc
}

func TestCompiler(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatalf("Failed to instanticate compiler")
	}
}

func TestCompile_RawText(t *testing.T) {
	compile(t, `Hello, World!`)
}

func TestCompile_LocalVar(t *testing.T) {
	compile(t, `[% s %]`)
}

func TestCompile_Wrapper(t *testing.T) {
	c := test.NewCtx(t)
	defer c.Cleanup()

	index := c.File("index.tx")
	index.WriteString(`[% WRAPPER "wrapper.tx" %]Hello[% END %]`)
	c.File("wrapper.tx").WriteString(`Hello, [% content %], Hello`)

	p := tterse.New()
	ast, err := p.Parse("index.tx", index.Read())
	if err != nil {
		t.Fatalf("Failed to parse template: %s", err)
	}

	comp := New()
	bc, err := comp.Compile(ast)
	if err != nil {
		t.Fatalf("Failed to compile ast: %s", err)
	}

	t.Logf("-> %+v", bc)
}
