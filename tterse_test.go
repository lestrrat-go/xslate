package xslate

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestTTerse_SimpleString(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`Hello, World!`, nil, `Hello, World!`)
	c.renderStringAndCompare(`    [%- "Hello, World!" %]`, nil, `Hello, World!`)
	c.renderStringAndCompare(`[% "Hello, World!" -%]    `, nil, `Hello, World!`)
}

func TestTTerse_SimpleHTMLString(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`<h1>Hello, World!</h1>`, nil, `<h1>Hello, World!</h1>`)
	c.renderStringAndCompare(`[% "<h1>Hello, World!</h1>" %]`, nil, `&lt;h1&gt;Hello, World!&lt;/h1&gt;`)
	c.renderStringAndCompare(`[% "<h1>Hello, World!</h1>" | mark_raw %]`, nil, `<h1>Hello, World!</h1>`)
}

func TestTTerse_Comment(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% # This is a comment %]Hello, World!`, nil, `Hello, World!`)
	c.renderStringAndCompare(`[% IF foo %]Hello, World![% END # DONE IF %]`, Vars{"foo": true}, `Hello, World!`)
}

func TestTTerse_Variable(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`Hello World, [% name %]!`, Vars{"name": "Bob"}, `Hello World, Bob!`)
	c.renderStringAndCompare(`[% x %]`, Vars{"x": uint32(1)}, `1`)
	c.renderStringAndCompare(`[% x %]`, Vars{"x": float64(0.32)}, `0.32`)

	now := time.Now()
	time.Sleep(time.Second)
	pattern := regexp.MustCompile(`\d+\.\d+`)
	c.renderStringAndCompare(`[% y.Sub(x).Seconds() %]`, Vars{"x": now, "y": time.Now()}, pattern)
}

func TestTTerse_MapVariable(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`Hello World, [% data.name %]!`, Vars{"data": map[string]string{"name": "Bob"}}, `Hello World, Bob!`)
}

func TestTTerse_ListVariableFunctions(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% list.size() %]`, Vars{"list": []int{0, 1, 2}}, `3`)
	c.renderStringAndCompare(`[% list.size() %]`, Vars{"list": []time.Time{}}, `0`)
}

func TestTTerse_ArrayVariableSlotAccess(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% SET l = [ 0..9 ] %][% l[0] %]..[% l[9] %]`, nil, `0..9`)
	c.renderStringAndCompare(`[% SET l = [ 0 .. 9 ] %][% l[0] %]..[% l[9] %]`, nil, `0..9`)
}

func TestTTerse_StructVariable(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`Hello World, [% data.name %]!`, Vars{"data": struct{ Name string }{"Bob"}}, `Hello World, Bob!`)
}

func TestTTerse_NestedStructVariable(t *testing.T) {
	inner := struct{ Name string }{"Bob"}
	outer := struct{ Inner struct{ Name string } }{inner}
	c := newTestCtx(t)
	c.renderStringAndCompare(`Hello World, [% data.inner.name %]!`, Vars{"data": outer}, `Hello World, Bob!`)
}

func TestTTerse_LocalVar(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% SET name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
	c.renderStringAndCompare(`[% name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
}

func TestTTerse_While(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% i = 0 %][% WHILE i < 10 %][% i %],[% CALL i += 1 %][% END %]`, Vars{"i": 0}, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestTTerse_Foreach(t *testing.T) {
	var list [10]int
	for i := 0; i < 10; i++ {
		list[i] = i
	}
	template := `[% FOREACH i IN list %][% i %],[% END %]`
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(template, Vars{"list": list}, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestTTerse_ForeachLoopVar(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	var template string
	template = `[% FOREACH i IN [0..9] %][% loop.index %],[% END %]`
	c.renderStringAndCompare(template, nil, `0,1,2,3,4,5,6,7,8,9,`)

	template = `[% FOREACH i IN [0..9] %][% loop.count %],[% END %]`
	c.renderStringAndCompare(template, nil, `1,2,3,4,5,6,7,8,9,10,`)

	template = `[% FOREACH i IN [0..9] %][% loop.size %],[% END %]`
	c.renderStringAndCompare(template, nil, `10,10,10,10,10,10,10,10,10,10,`)

	// need a way to elias is_first to IsFirst for compatibility
	// same for is_last and IsLast, peek_prev, peek_next, max_index
	template = `[% FOREACH i IN [0..9] %][% loop.IsFirst %],[% END %]`
	c.renderStringAndCompare(template, nil, `true,false,false,false,false,false,false,false,false,false,`)

	template = `[% FOREACH i IN [0..9] %][% loop.First %],[% END %]`
	c.renderStringAndCompare(template, nil, `true,false,false,false,false,false,false,false,false,false,`)

	template = `[% FOREACH i IN [0..9] %][% loop.IsLast %],[% END %]`
	c.renderStringAndCompare(template, nil, `false,false,false,false,false,false,false,false,false,true,`)

	template = `[% FOREACH i IN [0..9] %][% loop.MaxIndex %],[% END %]`
	c.renderStringAndCompare(template, nil, `9,9,9,9,9,9,9,9,9,9,`)
}

func TestTTerse_ForeachMakeArrayRange(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	template := `[% FOREACH i IN [0..9] %][% i %],[% END %]`
	c.renderStringAndCompare(template, nil, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestTTerse_ForeachMakeArrayList(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	template := `[% FOREACH i IN [0,1,2,3,4,5,6,7,8,9] %][% i %],[% END %]`
	c.renderStringAndCompare(template, nil, `0,1,2,3,4,5,6,7,8,9,`)

	template = `[% FOREACH i IN ["Alice", "Bob", "Charlie"] %][% i %],[% END %]`
	c.renderStringAndCompare(template, nil, `Alice,Bob,Charlie,`)
}

func TestTTerse_ForeachArrayInStruct(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	template := `[% FOREACH i IN foo.list %][% i %],[% END %]`
	vars := Vars{
		"foo": struct{ List []int }{[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}
	c.renderStringAndCompare(template, vars, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestTTerse_If(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	template := `[% IF (foo) %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": true}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": false}, ``)
}

func TestTTerse_IfElse(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	template := `[% IF (foo) %]Hello, World![% ELSE %]Goodbye, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": true}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": false}, `Goodbye, World!`)
}

func TestTTerse_Include(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("include/index.tx").WriteString(`[% INCLUDE "include/parts.tx" %]`)
	c.File("include/parts.tx").WriteString(`Hello, World! I'm included!`)
	c.File("include/include_var.tx").WriteString(`[% SET name = "include/parts.tx" %][% INCLUDE name %]`)

	tx := c.CreateTx()
	c.renderAndCompare(tx, "include/index.tx", nil, "Hello, World! I'm included!")
	c.renderAndCompare(tx, "include/include_var.tx", nil, "Hello, World! I'm included!")
}

func TestTTerse_IncludeWithArgs(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.File("include/index.tx").WriteString(`[% INCLUDE "include/parts.tx" WITH name = "Bob", foo = "Bar" %]`)
	c.File("include/parts.tx").WriteString(`Hello World, [% name %]!`)

	tx := c.CreateTx()
	c.renderAndCompare(tx, "include/index.tx", nil, "Hello World, Bob!")
}

func TestTTerse_Cache(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("test.tx").WriteString(`Hello World, [% name %]!`)

	tx := c.CreateTx()
	for i := range make([]struct{}, 10) {
		t0 := time.Now()
		c.renderAndCompare(tx, "test.tx", Vars{"name": "Alice"}, "Hello World, Alice!")
		t.Logf("Iteration %d took %s", i, time.Since(t0))
	}

	time.Sleep(5 * time.Millisecond)
	now := time.Now()
	err := os.Chtimes(c.Mkpath("test.tx"), now, now)
	if err != nil {
		t.Logf("Chtimes failed: %s", err)
	}

	for i := range make([]struct{}, 10) {
		t0 := time.Now()
		c.renderAndCompare(tx, "test.tx", Vars{"name": "Alice"}, "Hello World, Alice!")
		t.Logf("Iteration %d took %s", i, time.Since(t0))
	}
}

func TestTTerse_FunCallVariable(t *testing.T) {
	// Pass functions as variables. They are only available in the top most template,
	// and in successive templates, only if you pass it using WITH directive
	epoch := func() time.Time { return time.Unix(0, 0).In(time.UTC) }
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(`[% epoch() %]`, Vars{"epoch": epoch}, `1970-01-01 00:00:00 +0000 UTC`)

	// This one uses manual registering of functions, which are global
	x := func() {
		tx, err := New(Args{
			"Functions": Args{
				"epoch": epoch,
			},
		})
		if err != nil {
			t.Fatalf("Failed to create xslate: %s", err)
		}

		output, err := tx.RenderString(`[% epoch() %]`, nil)
		if err != nil {
			t.Fatalf("Failed to render: %s", err)
		}

		expected := `1970-01-01 00:00:00 +0000 UTC`
		if output != expected {
			t.Errorf("Expected '%s', got '%s'", expected, output)
		}
	}
	x()
}

func TestTTerse_MethodCallVariable(t *testing.T) {
	template := `[% t1.Before(t2) %]`

	c := newTestCtx(t)
	defer c.Cleanup()

	c.renderStringAndCompare(template, Vars{"t1": time.Unix(0, 0), "t2": time.Unix(100, 0)}, `true`)
}

func TestTTerse_Arithmetic(t *testing.T) {
	var template string

	c := newTestCtx(t)
	defer c.Cleanup()

	template = `[% 1 + 2 %]`
	c.renderStringAndCompare(template, nil, `3`)
	template = `[% 2 - 1 %]`
	c.renderStringAndCompare(template, nil, `1`)
	template = `[% 2 * 5 %]`
	c.renderStringAndCompare(template, nil, `10`)
	template = `[% 10 / 2 %]`
	c.renderStringAndCompare(template, nil, `5`)

	template = `[% (1 + 2) * 3 %]`
	c.renderStringAndCompare(template, nil, `9`)
	template = `[% 6 / ( 3 - 1) %]`
	c.renderStringAndCompare(template, nil, `3`)
	template = `[% 6 / ( 6 - (4 - 1)) %]`
	c.renderStringAndCompare(template, nil, `2`)
	template = `[% 6 / ( ( 4 - 1 ) - 1 ) %]`
	c.renderStringAndCompare(template, nil, `3`)

	template = `[% x = 0 %][% CALL x += 1 %][% CALL x += 1 %][% x %]`
	c.renderStringAndCompare(template, nil, `2`)
	template = `[% x = 2 %][% CALL x -= 1 %][% CALL x -= 1 %][% x %]`
	c.renderStringAndCompare(template, nil, `0`)
	template = `[% x = 2 %][% CALL x *= 2 %][% CALL x *= 2 %][% x %]`
	c.renderStringAndCompare(template, nil, `8`)
}

func TestTTerse_BooleanComparators(t *testing.T) {
	var template string

	c := newTestCtx(t)
	defer c.Cleanup()

	template = `[% IF foo > 10 %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": 20}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": 0}, ``)

	template = `[% IF foo < 10 %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": 20}, ``)
	c.renderStringAndCompare(template, Vars{"foo": 0}, `Hello, World!`)

	template = `[% IF foo == 10 %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": 10}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": 0}, ``)
	template = `[% IF foo == "bar" %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": "bar"}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": "baz"}, ``)
	template = `[% IF foo != "bar" %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": "bar"}, ``)
	c.renderStringAndCompare(template, Vars{"foo": "baz"}, `Hello, World!`)

	template = `[% IF foo != 10 %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": 0}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": 10}, ``)

	template = `[% IF foo.Size() < 10 %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": []int{}}, `Hello, World!`)

	// These exist solely for compatibility with perl5's Text::Xslate.
	// People using the go version are strongly discouraged from using
	// these operators
	template = `[% IF foo eq "bar" %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": "bar"}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": "baz"}, ``)

	template = `[% IF foo ne "bar" %]Hello, World![% END %]`
	c.renderStringAndCompare(template, Vars{"foo": "baz"}, `Hello, World!`)
	c.renderStringAndCompare(template, Vars{"foo": "bar"}, ``)
}

func TestTTerse_FilterHTML(t *testing.T) {
	template := `[% "<abc>" | html %]`

	c := newTestCtx(t)
	defer c.Cleanup()

	c.renderStringAndCompare(template, nil, `&lt;abc&gt;`)
}

func TestTTerse_FilterUri(t *testing.T) {
	template := `[% "日本語" | uri %]`

	c := newTestCtx(t)
	defer c.Cleanup()

	c.renderStringAndCompare(template, nil, `%E6%97%A5%E6%9C%AC%E8%AA%9E`)
}

func TestTTerse_Wrapper(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("wrapper/index.tx").WriteString(`[% WRAPPER "wrapper/wrapper.tx" %]<b>World</b>[% END %]`)
	c.File("wrapper/wrapper.tx").WriteString(`<html><body><h1>Hello [% content %] Bob!</h1></body></html>`)

	tx := c.CreateTx()
	c.renderAndCompare(tx, "wrapper/index.tx", nil, "<html><body><h1>Hello <b>World</b> Bob!</h1></body></html>")
}

func TestTTerse_WrapperWithArgs(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("wrapper/index.tx").WriteString(`[% WRAPPER "wrapper/wrapper.tx" WITH name = "Bob" %]Hello, Hello![% END %]`)
	c.File("wrapper/wrapper.tx").WriteString(`Hello World [% name %]! [% content %] Hello World [% name %]!`)

	tx := c.CreateTx()
	c.renderAndCompare(tx, "wrapper/index.tx", nil, "Hello World Bob! Hello, Hello! Hello World Bob!")
}

func TestTTerse_RenderInto(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("render_into/index.tx").WriteString(`Hello World, [% name %]!`)

	tx := c.CreateTx()

	b := &bytes.Buffer{}
	err := tx.RenderInto(b, "render_into/index.tx", Vars{"name": "Bob"})

	if err != nil {
		t.Fatalf("Failed to call RenderInto: %s", err)
	}

	if b.String() != "Hello World, Bob!" {
		t.Errorf("Expected 'Hello World, Bob!', got '%s'", b.String())
	}
}

func TestTTerse_Error(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()

	c.File("errors/index.tx").WriteString("Hello World,\n[% name ")

	tx := c.CreateTx()

	b := &bytes.Buffer{}
	err := tx.RenderInto(b, "errors/index.tx", Vars{"name": "Bob"})

	if err == nil {
		t.Fatalf("Expected error, got none")
	}

	if !strings.Contains(err.Error(), `Unexpected token found: Expected TagEnd, got Error ("unclosed tag") in errors/index.tx at line 2`) {
		t.Errorf("Could not find expected error string in '%s'", err)
	}
}

func TestTTerse_Macro(t *testing.T) {
	template := `
[%- MACRO repeat(text, count) BLOCK %]
[%- FOREACH i IN [1..count] %]
[% i %]: [% text %]
[%- END # FOREACH %]
[%- END -%]
[%- CALL repeat("Hello!", 10) -%]
  `

	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderStringAndCompare(template, nil, `
1: Hello
2: Hello
3: Hello
4: Hello
5: Hello
6: Hello
7: Hello
8: Hello
9: Hello
10: Hello`)
}

func TestTTerse_NilOnIfBlock(t *testing.T) {
	c := newTestCtx(t)
	defer c.Cleanup()
	c.renderString(`hello [% IF name %][% name %][% ELSE %]unknown[% END %]!`, Vars{"name":nil})
}
