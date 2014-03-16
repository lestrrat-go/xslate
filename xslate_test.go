package xslate

import (
  "fmt"
  "log"
  "os"
  "testing"

  "github.com/lestrrat/go-xslate/loader"
)

func ExampleXslate () {
  tx := New()
  tx.Loader, _ = loader.NewLoadFile(
    []string { "/path/to/templates" },
  )
  output, err := tx.Render("foo.tx", nil)
  if err != nil {
    log.Fatalf("Failed to render template: %s", err)
  }
  fmt.Fprintf(os.Stdout, output)
}

func executeAndCompare(t *testing.T, template string, vars Vars, expected string) {
  x := New()
  x.Flags |= DUMP_AST
  x.Flags |= DUMP_BYTECODE
  output, err := x.RenderString(template, vars)
  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  if output != expected {
    t.Errorf("Expected '%s', got '%s'", expected, output)
  }
}

func TestXslate_SimpleString(t *testing.T) {
  executeAndCompare(t, `Hello, World!`, nil, `Hello, World!`)
}

func TestXslate_Variable(t *testing.T) {
  executeAndCompare(t, `Hello World, [% name %]!`, Vars { "name": "Bob" }, `Hello World, Bob!`)
}

func TestXslate_MapVariable(t *testing.T) {
  executeAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": map[string]string { "name": "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_StructVariable(t *testing.T) {
  executeAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": struct { Name string } { "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_LocalVar(t *testing.T) {
  executeAndCompare(t, `[% SET name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
}

func TestXslate_Foreach(t *testing.T) {
  var list [10]int
  for i := 0; i < 10; i++ {
    list[i] = i
  }
  template := `[% FOREACH i IN list %][% i %],[% END %]`
  executeAndCompare(t, template, Vars { "list": list }, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_ForeachMakeArrayRange(t *testing.T) {
  template := `[% FOREACH i IN [0..9] %][% i %],[% END %]`
  executeAndCompare(t, template, nil, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_ForeachMakeArrayList(t *testing.T) {
  template := `[% FOREACH i IN [0,1,2,3,4,5,6,7,8,9] %][% i %],[% END %]`
  executeAndCompare(t, template, nil, `0,1,2,3,4,5,6,7,8,9,`)

  template = `[% FOREACH i IN ["Alice", "Bob", "Charlie"] %][% i %],[% END %]`
  executeAndCompare(t, template, nil, `Alice,Bob,Charlie,`)
}

func TestXslate_If(t *testing.T) {
  template := `[% IF (foo) %]Hello, World![% END %]`
  executeAndCompare(t, template, Vars { "foo": true }, `Hello, World!`)
  executeAndCompare(t, template, Vars { "foo": false }, ``)
}

func TestXslate_IfElse(t *testing.T) {
  template := `[% IF (foo) %]Hello, World![% ELSE %]Goodbye, World![% END %]`
  executeAndCompare(t, template, Vars { "foo": true }, `Hello, World!`)
  executeAndCompare(t, template, Vars { "foo": false }, `Goodbye, World!`)
}

