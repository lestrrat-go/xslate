/*
Package loader abstracts the top-level xslate package from the job of
loading the bytecode from a key value.
*/
package loader

import (
  "errors"
  "time"

  "github.com/lestrrat/go-xslate/vm"
)

// Mask... set of constants are used as flags to denote Debug modes.
// If you're using these you should be reading the source code
const (
  MaskDumpByteCode = 1 << iota
  MaskDumpAST
)

// Flags holds the flags to indicate if certain debug operations should be
// performed during load time
type Flags struct { flags int32 }

// NewFlags creates a new Flags struct initialized to 0
func NewFlags() *Flags {
  return &Flags { 0 }
}

// DumpAST sets the bitmask for DumpAST debug flag
func (f *Flags) DumpAST (b bool) {
  if b {
    f.flags |= MaskDumpAST
  } else {
    f.flags &= ^MaskDumpAST
  }
}

// DumpByteCode sets the bitmask for DumpByteCode debug flag
func (f *Flags) DumpByteCode (b bool) {
  if b {
    f.flags |= MaskDumpByteCode
  } else {
    f.flags &= ^MaskDumpByteCode
  }
}

// ShouldDumpAST returns true if the DumpAST debug flag is set
func (f *Flags) ShouldDumpAST() bool {
  return f.flags & MaskDumpAST == MaskDumpAST
}

// ShouldDumpByteCode returns true if the DumpByteCode debug flag is set
func (f Flags) ShouldDumpByteCode() bool {
  return f.flags & MaskDumpByteCode == 1
}

// DebugDumper defines interface that an object able to dump debug informatin 
// during load time must fulfill
type DebugDumper interface {
  DumpAST(bool)
  DumpByteCode(bool)
  ShouldDumpAST() bool
  ShouldDumpByteCode() bool
}

// ByteCodeLoader defines the interface for objects that can load
// ByteCode specified by a key
type ByteCodeLoader interface {
  DebugDumper
  LoadString(string, string) (*vm.ByteCode, error)
  Load(string) (*vm.ByteCode, error)
}

// TemplateFetcher defines the interface  for objects that can load
// TemplateSource specified by a key
type TemplateFetcher interface {
  FetchTemplate(string) (TemplateSource, error)
}

// TemplateSource is an abstraction over the actual template, which may live 
// on a file system, cache, database, whatever.
// It needs to be able to give us the actual template string AND its
// last modified time
type TemplateSource interface {
  LastModified() (time.Time, error)
  Bytes() ([]byte, error)
}

// ErrTemplateNotFound is returned whenever one of the loaders failed to
// find a suitable template
var ErrTemplateNotFound = errors.New("error: Specified template was not found")

