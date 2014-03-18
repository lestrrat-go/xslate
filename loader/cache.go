package loader

import(
  "encoding/gob"
  "errors"
  "os"
  "path/filepath"
  "strings"
  "github.com/lestrrat/go-xslate/compiler"
  "github.com/lestrrat/go-xslate/parser"
  "github.com/lestrrat/go-xslate/vm"
)

type Cache interface {
  Get(string) (*vm.ByteCode, error)
  Set(string, *vm.ByteCode) error
  Delete(string) error
}

type CachedByteCodeLoader struct {
  *StringByteCodeLoader // gives us LoadString
  Loader TemplateLoader
  Cache Cache
}

func NewCachedByteCodeLoader(
  cache Cache,
  loader TemplateLoader,
  parser parser.Parser,
  compiler compiler.Compiler,
) *CachedByteCodeLoader {
  return &CachedByteCodeLoader { 
    NewStringByteCodeLoader(parser, compiler),
    loader,
    cache,
  }
}

func (l *CachedByteCodeLoader) Load(key string) (*vm.ByteCode, error) {
  bc, err := l.Cache.Get(key)
  if err == nil {
    return bc, nil
  }

  template, err := l.Loader.Load(key)
  if err != nil {
    return nil, err
  }

  bc, err = l.LoadString(string(template))
  if err != nil {
    return nil, err
  }

  l.Cache.Set(key, bc)

  return bc, nil
}

type MemoryCache map[string]*vm.ByteCode

type FileCache struct {
  Dir string
}

func NewFileCache(dir string) (*FileCache, error) {
  f := &FileCache { dir }

STAT:
  fi, err := os.Stat(dir)
  if err != nil { // non-existing dir
    if err = os.MkdirAll(dir, 0777); err != nil {
      return nil, err
    }
    goto STAT
  }

  if ! fi.IsDir() {
    return nil, errors.New("Specified directory is not a directory!")
  }

  return f, nil
}

func (c *FileCache) GetCachePath(key string) string {
  // What's the best, portable way to remove make an absolute path into
  // a relative path?
  key = filepath.Clean(key)
  key = strings.TrimPrefix(key, "/")
  return filepath.Join(c.Dir, key)
}

func (c *FileCache) Get(key string) (*vm.ByteCode, error) {
  path := c.GetCachePath(key)

  // Need to avoid race condition
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var bc vm.ByteCode
  dec := gob.NewDecoder(file)
  if err = dec.Decode(&bc); err != nil {
    return nil, err
  }

  return &bc, nil
}

func (c *FileCache) Set(key string, bc *vm.ByteCode) error {
  path := c.GetCachePath(key)

  // Need to avoid race condition
  file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    return err
  }
  defer file.Close()

  enc := gob.NewEncoder(file)
  if err = enc.Encode(bc); err != nil {
    return err
  }

  return nil
}

func (c *FileCache) Delete(key string) error {
  return os.Remove(c.GetCachePath(key))
}
