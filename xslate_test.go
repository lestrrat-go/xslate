package xslate

import (
  "bytes"
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
  "reflect"
  "regexp"
  "testing"
  "time"
)

func createTx(path, cacheDir string, cacheLevel ...int) (*Xslate, error) {
  if len(cacheLevel) == 0 {
    cacheLevel = []int { 1 }
  }

  x, err := New(Args {
    // Optional. Currently only supports TTerse
    "Parser": Args {
      "Syntax": "TTerse",
    },
    // Compiler: DefaultCompiler, // don't need to specify
    "Loader": Args {
      "CacheDir": cacheDir,
      "CacheLevel": cacheLevel[0],
      "LoadPaths": []string { path },
    },
  })

  if err != nil {
    return nil, err
  }

  return x, nil
}

// Given key (relative path) => template content, creates physical files
// in a temporary location. The root directory, along with any error
// is returned
func generateTemplates(files map[string]string) (string, error) {
  baseDir, err := ioutil.TempDir("", "xslate-test-")
  if err != nil {
    panic("Failed to create temporary directory!")
  }

  for k, v := range files {
    fullpath := filepath.Join(baseDir, k)

    dir := filepath.Dir(fullpath)

STAT:
    fi, err := os.Stat(dir)
    if err != nil { // non-existent
      err = os.MkdirAll(dir, 0777)
      if err != nil {
        return "", err
      }

      goto STAT
    }

    if ! fi.IsDir() {
      return "", errors.New("error: Directory location already occupied by non-dir")
    }

    fh, err := os.OpenFile(fullpath, os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
      return "", err
    }
    defer fh.Close()

    _, err = fh.WriteString(v)
    if err != nil {
      return "", err
    }
  }

  return baseDir, nil
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

func renderStringAndCompare(t *testing.T, template string, vars Vars, expected interface {}) {
  x, _ := New()
  output, err := x.RenderString(template, vars)

  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  compareTemplateOutput(t, output, expected)
}

func renderAndCompare(t *testing.T, tx *Xslate, key string, vars Vars, expected string) {
  output, err := tx.Render(key, vars)
  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  compareTemplateOutput(t, output, expected)
}

func compareTemplateOutput(t *testing.T, output string, expected interface {}) {
  typ := reflect.TypeOf(expected)
  switch typ.Kind() {
  case reflect.String:
    if output != expected.(string) {
      t.Errorf("Expected '%s', got '%s'", expected, output)
    }
    return
  case reflect.Ptr:
    if typ.Elem().Name() == "Regexp" {
      if ! expected.(*regexp.Regexp).MatchString(output) {
        t.Errorf("Expected '%s', got '%s'", expected, output)
      }
      return
    }
  }
  t.Errorf("Unknown 'expected' type: %s", typ.Name())
}

func TestXslate_New_ParserSyntax(t *testing.T) {
  var err error
  _, err = New(Args { "Parser": Args { "Syntax": "Kolonish" } })
  if err == nil {
    t.Errorf("Expected Syntax: Kolonish to return an error, but got nothing")
  }

  _, err = New(Args { "Parser": Args { "Syntax": "TTerse" } })
  if err != nil {
    t.Errorf("Expected Syntax: TTerse to succeed, but got err: %s", err)
  }
}

func TestXslate_SimpleString(t *testing.T) {
  renderStringAndCompare(t, `Hello, World!`, nil, `Hello, World!`)
  renderStringAndCompare(t, `    [%- "Hello, World!" %]`, nil, `Hello, World!`)
  renderStringAndCompare(t, `[% "Hello, World!" -%]    `, nil, `Hello, World!`)
}


func TestXslate_SimpleHTMLString(t *testing.T) {
  renderStringAndCompare(t, `<h1>Hello, World!</h1>`, nil, `<h1>Hello, World!</h1>`)
}

func TestXslate_Comment(t *testing.T) {
  renderStringAndCompare(t, `[% # This is a comment %]Hello, World!`, nil, `Hello, World!`)
  renderStringAndCompare(t, `[% IF foo %]Hello, World![% END # DONE IF %]`, Vars { "foo": true }, `Hello, World!`)
}

func TestXslate_Variable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% name %]!`, Vars { "name": "Bob" }, `Hello World, Bob!`)
  renderStringAndCompare(t, `[% x %]`, Vars { "x": uint32(1) }, `1`)
  renderStringAndCompare(t, `[% x %]`, Vars { "x": float64(0.32) }, `0.32`)

  now := time.Now()
  time.Sleep(time.Second)
  pattern := regexp.MustCompile(`\d+\.\d+`)
  renderStringAndCompare(t, `[% y.Sub(x).Seconds() %]`, Vars { "x": now, "y": time.Now() }, pattern)
}

func TestXslate_MapVariable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": map[string]string { "name": "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_ListVariableFunctions(t *testing.T) {
  renderStringAndCompare(t, `[% list.size() %]`, Vars { "list": []int { 0, 1, 2 } }, `3`)
  renderStringAndCompare(t, `[% list.size() %]`, Vars { "list": []time.Time { } }, `0`)
}

func TestXslate_ArrayVariableSlotAccess(t *testing.T) {
  renderStringAndCompare(t, `[% SET l = [ 0..9 ] %][% l[0] %]..[% l[9] %]`, nil, `0..9`)
}

func TestXslate_StructVariable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": struct { Name string } { "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_NestedStructVariable(t *testing.T) {
  inner := struct { Name string } { "Bob" }
  outer := struct { Inner struct { Name string } } { inner }
  renderStringAndCompare(t, `Hello World, [% data.inner.name %]!`, Vars { "data": outer }, `Hello World, Bob!`)
}

func TestXslate_LocalVar(t *testing.T) {
  renderStringAndCompare(t, `[% SET name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
  renderStringAndCompare(t, `[% name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
}

func TestXslate_While(t *testing.T) {
  renderStringAndCompare(t, `[% i = 0 %][% WHILE i < 10 %][% i %],[% CALL i += 1 %][% END %]`, Vars { "i": 0 }, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_Foreach(t *testing.T) {
  var list [10]int
  for i := 0; i < 10; i++ {
    list[i] = i
  }
  template := `[% FOREACH i IN list %][% i %],[% END %]`
  renderStringAndCompare(t, template, Vars { "list": list }, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_ForeachLoopVar(t *testing.T) {
  var template string
  template = `[% FOREACH i IN [0..9] %][% loop.index %],[% END %]`
  renderStringAndCompare(t, template, nil, `0,1,2,3,4,5,6,7,8,9,`)

  template = `[% FOREACH i IN [0..9] %][% loop.count %],[% END %]`
  renderStringAndCompare(t, template, nil, `1,2,3,4,5,6,7,8,9,10,`)

  template = `[% FOREACH i IN [0..9] %][% loop.size %],[% END %]`
  renderStringAndCompare(t, template, nil, `10,10,10,10,10,10,10,10,10,10,`)

  // need a way to elias is_first to IsFirst for compatibility
  // same for is_last and IsLast, peek_prev, peek_next, max_index
  template = `[% FOREACH i IN [0..9] %][% loop.IsFirst %],[% END %]`
  renderStringAndCompare(t, template, nil, `true,false,false,false,false,false,false,false,false,false,`)

  template = `[% FOREACH i IN [0..9] %][% loop.First %],[% END %]`
  renderStringAndCompare(t, template, nil, `true,false,false,false,false,false,false,false,false,false,`)

  template = `[% FOREACH i IN [0..9] %][% loop.IsLast %],[% END %]`
  renderStringAndCompare(t, template, nil, `false,false,false,false,false,false,false,false,false,true,`)

  template = `[% FOREACH i IN [0..9] %][% loop.MaxIndex %],[% END %]`
  renderStringAndCompare(t, template, nil, `9,9,9,9,9,9,9,9,9,9,`)
}

func TestXslate_ForeachMakeArrayRange(t *testing.T) {
  template := `[% FOREACH i IN [0..9] %][% i %],[% END %]`
  renderStringAndCompare(t, template, nil, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_ForeachMakeArrayList(t *testing.T) {
  template := `[% FOREACH i IN [0,1,2,3,4,5,6,7,8,9] %][% i %],[% END %]`
  renderStringAndCompare(t, template, nil, `0,1,2,3,4,5,6,7,8,9,`)

  template = `[% FOREACH i IN ["Alice", "Bob", "Charlie"] %][% i %],[% END %]`
  renderStringAndCompare(t, template, nil, `Alice,Bob,Charlie,`)
}

func TestXslate_ForeachArrayInStruct(t *testing.T) {
  template := `[% FOREACH i IN foo.list %][% i %],[% END %]`
  vars := Vars {
    "foo": struct { List []int } { []int{ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 } },
  }
  renderStringAndCompare(t, template, vars, `0,1,2,3,4,5,6,7,8,9,`)
}

func TestXslate_If(t *testing.T) {
  template := `[% IF (foo) %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": true }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": false }, ``)
}

func TestXslate_IfElse(t *testing.T) {
  template := `[% IF (foo) %]Hello, World![% ELSE %]Goodbye, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": true }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": false }, `Goodbye, World!`)
}

func TestXslate_Include(t *testing.T) {
  files := map[string]string {
    "include/index.tx": `[% INCLUDE "include/parts.tx" %]`,
    "include/parts.tx": `Hello, World! I'm included!`,
    "include/include_var.tx": `[% SET name = "include/parts.tx" %][% INCLUDE name %]`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }
  renderAndCompare(t, tx, "include/index.tx", nil, "Hello, World! I'm included!")
  renderAndCompare(t, tx, "include/include_var.tx", nil, "Hello, World! I'm included!")
}

func TestXslate_IncludeWithArgs(t *testing.T) {
  files := map[string]string {
    "include/index.tx": `[% INCLUDE "include/parts.tx" WITH name = "Bob", foo = "Bar" %]`,
    "include/parts.tx": `Hello World, [% name %]!`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }
  renderAndCompare(t, tx, "include/index.tx", nil, "Hello World, Bob!")
}

func TestXslate_Cache(t *testing.T) {
  files := map[string]string {
    "test.tx": `Hello World, [% name %]!`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }

  for i := range make([]struct {}, 10) {
    t0 := time.Now()
    renderAndCompare(t, tx, "test.tx", Vars { "name": "Alice" }, "Hello World, Alice!")
    t.Logf("Iteration %d took %s", i, time.Since(t0))
  }

  time.Sleep(5 * time.Millisecond)
  now := time.Now()
  err = os.Chtimes(filepath.Join(root, "test.tx"), now, now)
  if err != nil {
    t.Logf("Chtimes failed: %s", err)
  }

  for i := range make([]struct {}, 10) {
    t0 := time.Now()
    renderAndCompare(t, tx, "test.tx", Vars { "name": "Alice" }, "Hello World, Alice!")
    t.Logf("Iteration %d took %s", i, time.Since(t0))
  }
}

func TestXslate_FunCallVariable(t *testing.T) {
  template := `[% epoch() %]`
  renderStringAndCompare(t, template, Vars { "epoch": func() time.Time { return time.Unix(0, 0).In(time.UTC) } }, `1970-01-01 00:00:00 +0000 UTC`)
}

func TestXslate_MethodCallVariable(t *testing.T) {
  template := `[% t1.Before(t2) %]`
  renderStringAndCompare(t, template, Vars { "t1": time.Unix(0, 0), "t2": time.Unix(100, 0) }, `true`)
}

func TestXslate_Arithmetic(t *testing.T) {
  var template string

  template = `[% 1 + 2 %]`
  renderStringAndCompare(t, template, nil, `3`)
  template = `[% 2 - 1 %]`
  renderStringAndCompare(t, template, nil, `1`)
  template = `[% 2 * 5 %]`
  renderStringAndCompare(t, template, nil, `10`)
  template = `[% 10 / 2 %]`
  renderStringAndCompare(t, template, nil, `5`)

  template = `[% (1 + 2) * 3 %]`
  renderStringAndCompare(t, template, nil, `9`)
  template = `[% 6 / ( 3 - 1) %]`
  renderStringAndCompare(t, template, nil, `3`)
  template = `[% 6 / ( 6 - (4 - 1)) %]`
  renderStringAndCompare(t, template, nil, `2`)
  template = `[% 6 / ( ( 4 - 1 ) - 1 ) %]`
  renderStringAndCompare(t, template, nil, `3`)

  template = `[% x = 0 %][% CALL x += 1 %][% CALL x += 1 %][% x %]`
  renderStringAndCompare(t, template, nil, `2`)
  template = `[% x = 2 %][% CALL x -= 1 %][% CALL x -= 1 %][% x %]`
  renderStringAndCompare(t, template, nil, `0`)
  template = `[% x = 2 %][% CALL x *= 2 %][% CALL x *= 2 %][% x %]`
  renderStringAndCompare(t, template, nil, `8`)
}

func TestXslate_BooleanComparators(t *testing.T) {
  var template string

  template = `[% IF foo > 10 %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": 20 }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": 0 }, ``)

  template = `[% IF foo < 10 %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": 20 }, ``)
  renderStringAndCompare(t, template, Vars { "foo": 0 }, `Hello, World!`)

  template = `[% IF foo == 10 %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": 10 }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": 0 }, ``)
  template = `[% IF foo == "bar" %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": "bar" }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": "baz" }, ``)
  template = `[% IF foo != "bar" %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": "bar" }, ``)
  renderStringAndCompare(t, template, Vars { "foo": "baz" }, `Hello, World!`)

  template = `[% IF foo != 10 %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": 0 }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": 10 }, ``)

  template = `[% IF foo.Size() < 10 %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": []int {} }, `Hello, World!`)

  // These exist solely for compatibility with perl5's Text::Xslate.
  // People using the go version are strongly discouraged from using
  // these operators
  template = `[% IF foo eq "bar" %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": "bar" }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": "baz" }, ``)

  template = `[% IF foo ne "bar" %]Hello, World![% END %]`
  renderStringAndCompare(t, template, Vars { "foo": "baz" }, `Hello, World!`)
  renderStringAndCompare(t, template, Vars { "foo": "bar" }, ``)
}

func TestXslate_FilterHTML(t *testing.T) {
  template := `[% "<abc>" | html %]`
  renderStringAndCompare(t, template, nil, `&lt;abc&gt;`)
}

func TestXslate_FilterUri(t *testing.T) {
  template := `[% "日本語" | uri %]`
  renderStringAndCompare(t, template, nil, `%E6%97%A5%E6%9C%AC%E8%AA%9E`)
}

func TestXslate_Wrapper(t *testing.T) {
  files := map[string]string {
    "wrapper/index.tx": `[% WRAPPER "wrapper/wrapper.tx" %]<b>World</b>[% END %]`,
    "wrapper/wrapper.tx": `<html><body><h1>Hello [% content %] Bob!</h1></body></html>`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }
  renderAndCompare(t, tx, "wrapper/index.tx", nil, "<html><body><h1>Hello <b>World</b> Bob!</h1></body></html>")
}

func TestXslate_WrapperWithArgs(t *testing.T) {
  files := map[string]string {
    "wrapper/index.tx": `[% WRAPPER "wrapper/wrapper.tx" WITH name = "Bob" %]Hello, Hello![% END %]`,
    "wrapper/wrapper.tx": `Hello World [% name %]! [% content %] Hello World [% name %]!`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }
  renderAndCompare(t, tx, "wrapper/index.tx", nil, "Hello World Bob! Hello, Hello! Hello World Bob!")
}

func TestXslate_RenderInto(t *testing.T) {
  files := map[string]string {
    "render_into/index.tx": `Hello World, [% name %]!`,
  }

  root, err := generateTemplates(files)
  if err != nil {
    t.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"))
  if err != nil {
    t.Fatalf("Failed to create xslate instance: %s", err)
  }

  b := &bytes.Buffer {}
  err = tx.RenderInto(b, "render_into/index.tx", Vars { "name": "Bob" })

  if err != nil {
    t.Fatalf("Failed to call RenderInto: %s", err)
  }

  if b.String() != "Hello World, Bob!" {
    t.Errorf("Expected 'Hello World, Bob!', got '%s'", b.String())
  }
}

