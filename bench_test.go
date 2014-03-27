package xslate

import (
  "bytes"
  tt "text/template"
  ht "html/template"
  "os"
  "path/filepath"
  "testing"
)

func BenchmarkXslateHelloWorld(b *testing.B) {
  files := map[string]string {
    "xslate/hello.tx": `Hello World, [% name %]!`,
  }
  root, err := generateTemplates(files)
  if err != nil {
    b.Fatalf("Failed to create template files: %s", err)
  }
  defer os.RemoveAll(root)

  tx, err := createTx(root, filepath.Join(root, "cache"), 2)
  if err != nil {
    b.Fatalf("Failed to create xslate instance: %s", err)
  }

  vars := Vars { "name": "Bob" }
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    buf := &bytes.Buffer {}
    tx.RenderInto(buf, "xslate/hello.tx", vars)
  }
}

func BenchmarkHTMLTemplateHelloWorld(b *testing.B) {
  t, err := ht.New("hello").Parse(`{{define "T"}}Hello World, {{.}}!{{end}}`)
  if err != nil {
    b.Fatalf("Failed to parse template: %s", err)
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    buf := &bytes.Buffer {}
    t.ExecuteTemplate(buf, "T", "Bob")
  }
}

func BenchmarkTextTemplateHelloWorld(b *testing.B) {
  t, err := tt.New("hello").Parse(`{{define "T"}}Hello World, {{.}}!{{end}}`)
  if err != nil {
    b.Fatalf("Failed to parse template: %s", err)
  }

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    buf := &bytes.Buffer {}
    t.ExecuteTemplate(buf, "T", "Bob")
  }
}
