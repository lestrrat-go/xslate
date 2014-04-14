package test

import (
  "io/ioutil"
  "os"
  "path"
  "path/filepath"
)

type Ctx struct {
  Tester
  BaseDir string
}

type Tester interface {
  Errorf(string, ...interface{})
  Fatalf(string, ...interface {})
  Logf(string, ...interface{})
}

/*

NewCtx creates a new Context

  ctx := test.NewCtx()
  defer ctx.Clean()
  ctx.File(...).WriteSTring

*/
func NewCtx(t Tester) *Ctx {
  dir, err := ioutil.TempDir("", "xslate-test-")
  if err != nil {
    panic("Failed to create temporary directory!")
  }

  return &Ctx { t, dir }
}

func (c *Ctx) Cleanup() {
  os.RemoveAll(c.BaseDir)
}

type File struct {
  *Ctx
  Path string
}

func (c *Ctx) File(name string) *File {
  return &File {c, name}
}

func (c *Ctx) Mkpath(name string) string {
  return filepath.Join(c.BaseDir, name)
}

func (f *File) FullPath() string {
  return f.Mkpath(f.Path)
}

func (f *File) Mkdir() {
  fullpath := f.FullPath()
  dir := path.Dir(fullpath)

  _, err := os.Stat(dir)
  if err != nil { // non-existent
    err = os.MkdirAll(dir, 0777)
    if err != nil {
      f.Fatalf("error: Mkdir %s failed: %s", dir, err)
    }
  }
}

func (f *File) WriteString(body string) {
  f.Mkdir()

  fullpath := f.FullPath()
  fh, err := os.OpenFile(fullpath, os.O_CREATE|os.O_WRONLY, 0666)
  if err != nil {
    f.Fatalf("error: Failed to open file %s for writing: %s", fullpath, err)
  }
  defer fh.Close()

  _, err = fh.WriteString(body)
  if err != nil {
    f.Fatalf("error: Failed to write to file %s: %s", fullpath, err)
  }
}

func (f *File) Read() ([]byte) {
  fh, err := os.Open(f.FullPath())
  if err != nil {
    f.Fatalf("error: Failed to open file %s for reading: %s", f.FullPath(), err)
  }
  defer fh.Close()

  buf, err := ioutil.ReadAll(fh)
  if err != nil {
    f.Fatalf("error: Failed to read from file %s: %s", f.FullPath(), err)
  }
  return buf
}
