package xslate

import (
  "bytes"
  "github.com/lestrrat/go-xslate/test"
  tt "text/template"
  ht "html/template"
  "testing"
)

func BenchmarkXslateHelloWorld(b *testing.B) {
  c := test.NewCtx(b)
  defer c.Cleanup()

  c.File("xslate/hello.tx").WriteString(`Hello World, [% name %]!`)

  tx, err := createTx(c.BaseDir, c.Mkpath("cache"), 2)
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
