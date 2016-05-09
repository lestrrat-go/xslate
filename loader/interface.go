package loader

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"

	"github.com/lestrrat/go-xslate/compiler"
	"github.com/lestrrat/go-xslate/vm"
	"github.com/lestrrat/go-xslate/parser"
)

// CacheStrategy specifies how the cache should be checked
type CacheStrategy int

const (
	// CacheNone flag specifies that cache checking and setting hould be skipped
	CacheNone CacheStrategy = iota
	// CacheVerify flag specifies that cached ByteCode generation time should be
	// verified against the source's last modified time. If new, the source is
	// re-parsed and re-compiled even on a cache hit.
	CacheVerify
	// CacheNoVerify flag specifies that if we have a cache hit, the ByteCode
	// is not verified against the source. If there's a cache hit, it is
	// used regardless of updates to the original template on file system
	CacheNoVerify
)

// CacheEntity contains all the othings required to perform calculations
// necessary to validate a template
type CacheEntity struct {
	ByteCode *vm.ByteCode
	Source   TemplateSource
}

// Cache defines the interface for things that can cache generated ByteCode
type Cache interface {
	Get(string) (*CacheEntity, error)
	Set(string, *CacheEntity) error
	Delete(string) error
}

// CachedByteCodeLoader is the default ByteCodeLoader that loads templates
// from the file system and caches in the file system, too
type CachedByteCodeLoader struct {
	*StringByteCodeLoader // gives us LoadString
	*ReaderByteCodeLoader // gives us LoadReader
	Fetcher               TemplateFetcher
	Caches                []Cache
	CacheLevel            CacheStrategy
}

// FileCache is Cache implementation that stores caches in the file system
type FileCache struct {
	Dir string
}

// MemoryCache is what's used store cached ByteCode in memory for maximum
// speed. As of this writing this cache never freed. We may need to
// introduce LRU in the future
type MemoryCache map[string]*CacheEntity

// FileTemplateFetcher is a TemplateFetcher that loads template strings
// in the file system.
type FileTemplateFetcher struct {
	Paths []string
}

// NewFileSource creates a new FileSource
func NewFileSource(path string) *FileSource {
	return &FileSource{path, time.Time{}, nil}
}

// FileSource is a TemplateSource variant that holds template information
// in a file.
type FileSource struct {
	Path           string
	LastStat       time.Time
	LastStatResult os.FileInfo
}

// HTTPTemplateFetcher is a proof of concept loader that fetches templates
// from external http servers. Probably not a good thing to use in
// your production environment
type HTTPTemplateFetcher struct {
	URLs []string
}

// HTTPSource represents a template source fetched via HTTP
type HTTPSource struct {
	Buffer           *bytes.Buffer
	LastModifiedTime time.Time
}

// ReaderByteCodeLoader is a fancy name for objects that can "given a template
// string, parse and compile it". This is one of the most common operations
// that users want to do, but it needs to be separate from other loaders
// because there's no sane way to cache intermediate results, and therefore
// has significant performance penalty
type ReaderByteCodeLoader struct {
	*Flags
	Parser   parser.Parser
	Compiler compiler.Compiler
}

// Mask... set of constants are used as flags to denote Debug modes.
// If you're using these you should be reading the source code
const (
	MaskDumpByteCode = 1 << iota
	MaskDumpAST
)

// Flags holds the flags to indicate if certain debug operations should be
// performed during load time
type Flags struct{ flags int32 }

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
	Reader() (io.Reader, error)
}

// ErrTemplateNotFound is returned whenever one of the loaders failed to
// find a suitable template
var ErrTemplateNotFound = errors.New("error: Specified template was not found")

// StringByteCodeLoader is a fancy name for objects that can "given a template
// string, parse and compile it". This is one of the most common operations
// that users want to do, but it needs to be separate from other loaders
// because there's no sane way to cache intermediate results, and therefore
// has significant performance penalty
type StringByteCodeLoader struct {
	*Flags
	Parser   parser.Parser
	Compiler compiler.Compiler
}