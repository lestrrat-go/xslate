/*
Package loader abstracts the top-level xslate package from the job of
loading the bytecode from a key value.
*/
package loader

// NewFlags creates a new Flags struct initialized to 0
func NewFlags() *Flags {
	return &Flags{0}
}

// DumpAST sets the bitmask for DumpAST debug flag
func (f *Flags) DumpAST(b bool) {
	if b {
		f.flags |= MaskDumpAST
	} else {
		f.flags &= ^MaskDumpAST
	}
}

// DumpByteCode sets the bitmask for DumpByteCode debug flag
func (f *Flags) DumpByteCode(b bool) {
	if b {
		f.flags |= MaskDumpByteCode
	} else {
		f.flags &= ^MaskDumpByteCode
	}
}

// ShouldDumpAST returns true if the DumpAST debug flag is set
func (f *Flags) ShouldDumpAST() bool {
	return f.flags&MaskDumpAST == MaskDumpAST
}

// ShouldDumpByteCode returns true if the DumpByteCode debug flag is set
func (f Flags) ShouldDumpByteCode() bool {
	return f.flags&MaskDumpByteCode == 1
}
