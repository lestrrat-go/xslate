package xslate

import (
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
  "testing"
  "time"
)

func createTx(path, cacheDir string) (*Xslate, error) {
  x, err := New(Args {
    // Compiler: DefaultCompiler, // don't need to specify
    // Parser: DefaultParser, // don't need to specify
    "Loader": Args {
      "CacheDir": cacheDir,
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
      return "", errors.New("Directory location already occupied by non-dir")
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

func renderStringAndCompare(t *testing.T, template string, vars Vars, expected string) {
  x, _ := New()

  x.DumpAST(true)
  x.DumpByteCode(true)

  output, err := x.RenderString(template, vars)

  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  compareTemplateOutput(t, output, expected)
}

func renderAndCompare(t *testing.T, tx *Xslate, key string, vars Vars, expected string) {
  tx.DumpAST(true)
  tx.DumpByteCode(true)

  output, err := tx.Render(key, vars)
  if err != nil {
    t.Fatalf("Failed to render template: %s", err)
  }
  compareTemplateOutput(t, output, expected)
}

func compareTemplateOutput(t *testing.T, output, expected string) {
  if output != expected {
    t.Errorf("Expected '%s', got '%s'", expected, output)
  }
}

func TestXslate_SimpleString(t *testing.T) {
  renderStringAndCompare(t, `Hello, World!`, nil, `Hello, World!`)
}

func TestXslate_Variable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% name %]!`, Vars { "name": "Bob" }, `Hello World, Bob!`)
}

func TestXslate_MapVariable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": map[string]string { "name": "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_StructVariable(t *testing.T) {
  renderStringAndCompare(t, `Hello World, [% data.name %]!`, Vars { "data": struct { Name string } { "Bob" } }, `Hello World, Bob!`)
}

func TestXslate_LocalVar(t *testing.T) {
  renderStringAndCompare(t, `[% SET name = "Bob" %]Hello World, [% name %]!`, nil, `Hello World, Bob!`)
}

func TestXslate_Foreach(t *testing.T) {
  var list [10]int
  for i := 0; i < 10; i++ {
    list[i] = i
  }
  template := `[% FOREACH i IN list %][% i %],[% END %]`
  renderStringAndCompare(t, template, Vars { "list": list }, `0,1,2,3,4,5,6,7,8,9,`)
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

  time.Sleep(time.Second)
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

