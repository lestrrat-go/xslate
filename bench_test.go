package xslate

import (
  "bytes"
  "html/template"
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
  for i := 0; i < b.N; i++ {
    tx.Render("xslate/hello.tx", vars)
  }
}

func BenchmarkHTMLTemplateHelloWorld(b *testing.B) {
  t, err := template.New("hello").Parse(`{{define "T"}}Hello World, {{.}}!{{end}}`)
  if err != nil {
    b.Fatalf("Failed to parse template: %s", err)
  }

  for i := 0; i < b.N; i++ {
    buf := &bytes.Buffer {}
    t.ExecuteTemplate(buf, "T", "Bob")
  }
}