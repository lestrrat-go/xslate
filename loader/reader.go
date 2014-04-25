package loader

import (
	"fmt"
	"github.com/lestrrat/go-xslate/compiler"
	"github.com/lestrrat/go-xslate/parser"
	"github.com/lestrrat/go-xslate/vm"
	"io"
	"os"
)

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

// NewReaderByteCodeLoader creates a new object
func NewReaderByteCodeLoader(p parser.Parser, c compiler.Compiler) *ReaderByteCodeLoader {
	return &ReaderByteCodeLoader{NewFlags(), p, c}
}

// LoadReader takes a io.Reader and compiles it into vm.ByteCode
func (l *ReaderByteCodeLoader) LoadReader(name string, rdr io.Reader) (*vm.ByteCode, error) {
	ast, err := l.Parser.ParseReader(name, rdr)
	if err != nil {
		return nil, err
	}

	if l.ShouldDumpAST() {
		fmt.Fprintf(os.Stderr, "AST:\n%s\n", ast)
	}

	bc, err := l.Compiler.Compile(ast)
	if err != nil {
		return nil, err
	}

	return bc, nil
}
