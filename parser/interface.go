package parser

import (
	"io"
	"time"

	"github.com/lestrrat/go-lex"
	"github.com/lestrrat/go-xslate/node"
	"github.com/lestrrat/go-xslate/util"
)

const (
	ItemError lex.ItemType = lex.ItemDefaultMax + 1 + iota
	ItemEOF
	ItemRawString
	ItemComment
	ItemNumber
	ItemComplex
	ItemChar
	ItemSpace
	ItemTagStart
	ItemTagEnd
	ItemSymbol
	ItemIdentifier
	ItemDoubleQuotedString
	ItemSingleQuotedString
	ItemBool
	ItemField
	ItemComma
	ItemOpenParen          // '('
	ItemCloseParen         // ')'
	ItemOpenSquareBracket  // '['
	ItemCloseSquareBracket // ']'
	ItemPeriod             // '.'
	ItemKeyword            // Delimiter
	ItemCall               // CALL
	ItemGet                // GET
	ItemSet                // SET
	ItemMacro              // MACRO
	ItemBlock              // BLOCK
	ItemForeach            // FOREACH
	ItemWhile              // WHILE
	ItemIn                 // IN
	ItemInclude            // INCLUDE
	ItemWith               // WITH
	ItemIf                 // IF
	ItemElse               // ELSE
	ItemElseIf             // ELSIF
	ItemUnless             // UNLESS
	ItemSwitch             // SWITCH
	ItemCase               // CASE
	ItemWrapper            // WRAPPER
	ItemDefault            // DEFAULT
	ItemEnd                // END
	ItemOperator           // Delimiter
	ItemRange              // ..
	ItemEquals             // ==
	ItemNotEquals          // !=
	ItemGT                 // >
	ItemLT                 // <
	ItemCmp                // <=>
	ItemLE                 // <=
	ItemGE                 // >=
	ItemShiftLeft          // <<
	ItemShiftRight         // >>
	ItemAssignAdd          // +=
	ItemAssignSub          // -=
	ItemAssignMul          // *=
	ItemAssignDiv          // /=
	ItemAssignMod          // %=
	ItemAnd                // &&
	ItemOr                 // ||
	ItemFatComma           // =>
	ItemIncr               // ++
	ItemDecr               // --
	ItemPlus
	ItemMinus
	ItemAsterisk
	ItemSlash
	ItemVerticalSlash
	ItemMod
	ItemAssign // =

	DefaultItemTypeMax
)

// AST is represents the syntax tree for an Xslate template
type AST struct {
	Name      string         // name of the template
	ParseName string         // name of the top-level template during parsing
	Root      *node.ListNode // root of the tree
	Timestamp time.Time      // last-modified date of this template
	text      string
}

type Builder struct {
}

// Frame is the frame struct used during parsing, which has a bit of
// extension over the common Frame struct.
type Frame struct {
	*util.Frame
	Node node.Appender

	// This contains names of local variables, mapped to their
	// respective location in the framestack
	LvarNames map[string]int
}

type Lexer struct {
	lex.Lexer
	tagStart string
	tagEnd   string
	symbols  *LexSymbolSet
}

// LexSymbol holds the pre-defined symbols to be lexed
type LexSymbol struct {
	Name     string
	Type     lex.ItemType
	Priority float32
}

// LexSymbolList a list of LexSymbols. Normally you do not need to use it.
// This is mainly only useful for sorting LexSymbols
type LexSymbolList []LexSymbol

// LexSymbolSorter sorts a list of LexSymbols by priority
type LexSymbolSorter struct {
	list LexSymbolList
}

// LexSymbolSet is the container for symbols.
type LexSymbolSet struct {
	Map        map[string]LexSymbol
	SortedList LexSymbolList
}

// Parser defines the interface for Xslate parsers
type Parser interface {
	Parse(string, []byte) (*AST, error)
	ParseString(string, string) (*AST, error)
	ParseReader(string, io.Reader) (*AST, error)
}
