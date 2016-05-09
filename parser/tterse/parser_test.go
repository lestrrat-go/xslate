package tterse

import (
	"testing"

	"github.com/lestrrat/go-xslate/node"
	"github.com/lestrrat/go-xslate/parser"
)

func parse(t *testing.T, tmpl string) *parser.AST {
	p := New()
	ast, err := p.ParseString(tmpl, tmpl)
	if err != nil {
		t.Fatalf("Failed to parse template: %s", err)
	}
	return ast
}

func matchNodeTypes(t *testing.T, ast *parser.AST, expected []node.NodeType) {
	i := 0
	for n := range ast.Visit() {
		t.Logf("n -> %s", n.Type())

		if len(expected) <= i {
			t.Fatalf("Got extra nodes after %d nodes", i)
		}

		if n.Type() != expected[i] {
			t.Fatalf("Expected node type %s, got %s", expected[i], n.Type())
		}
		i++
	}

	if i < len(expected) {
		t.Fatalf("Expected %d nodes, but only got %d", len(expected), i)
	}
}

func TestRawString(t *testing.T) {
	tmpl := `Hello, World!`
	ast := parse(t, tmpl)

	// Expect nodes to be in this order:
	expected := []node.NodeType{
		node.NodeRoot,
		node.NodePrintRaw,
		node.NodeText,
	}
	matchNodeTypes(t, ast, expected)
}

func TestGetLocalVariable(t *testing.T) {
	tmpl := `[% SET name = "Bob" %]Hello World, [% name %]`
	ast := parse(t, tmpl)

	expected := []node.NodeType{
		node.NodeRoot,
		node.NodeAssignment,
		node.NodeLocalVar,
		node.NodeText,
		node.NodePrintRaw,
		node.NodeText,
		node.NodePrint,
		node.NodeLocalVar,
	}
	matchNodeTypes(t, ast, expected)
}

func TestForeachLoop(t *testing.T) {
	tmpl := `[% FOREACH x IN list %]Hello World, [% x %][% END %]`
	ast := parse(t, tmpl)
	expected := []node.NodeType{
		node.NodeRoot,
		node.NodeForeach,
		node.NodePrintRaw,
		node.NodeText,
		node.NodePrint,
		node.NodeLocalVar,
	}
	matchNodeTypes(t, ast, expected)
}

func TestBasic(t *testing.T) {
	tmpl := `
[% WRAPPER "hoge.tx" WITH foo = "bar" %]
[% FOREACH x IN list %]
[% loop.index %]. x is [% x %]
[% END %]
[% END %]
`
	p := New()
	ast, err := p.ParseString(tmpl, tmpl)
	if err != nil {
		t.Errorf("Error during parse: %s", err)
	}

	if len(ast.Root.Nodes) == 1 {
		t.Errorf("Expected Root node to have 1 child, got %d", len(ast.Root.Nodes))
	}
}

func TestSimpleAssign(t *testing.T) {
	tmpl := `[% SET s = 1 %][% s %]`
	ast := parse(t, tmpl)

	expected := []node.NodeType{
		node.NodeRoot,
		node.NodeAssignment,
		node.NodeLocalVar,
		node.NodeInt,
		node.NodePrint,
		node.NodeLocalVar,
	}

	matchNodeTypes(t, ast, expected)
}
