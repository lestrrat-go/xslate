package parser

/*

Lexer for TTerse, based on http://golang.org/src/pkg/text/template/parse/lex.go

Anything up to a tagStart('[%') is treated as RawText, and therefore does not
need any real lexing.

Once tagStart is found, real lexing starts.

*/

import (
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/lestrrat-go/lex"
)

func init() {
	lex.TypeNames[ItemError] = "Error"
	lex.TypeNames[ItemRawString] = "RawString"
	lex.TypeNames[ItemEOF] = "EOF"
	lex.TypeNames[ItemComment] = "Comment"
	lex.TypeNames[ItemComplex] = "Complex" // may not need this
	lex.TypeNames[ItemChar] = "Char"
	lex.TypeNames[ItemSpace] = "Space"
	lex.TypeNames[ItemNumber] = "Number"
	lex.TypeNames[ItemSymbol] = "Symbol"
	lex.TypeNames[ItemIdentifier] = "Identifier"
	lex.TypeNames[ItemTagStart] = "TagStart"
	lex.TypeNames[ItemTagEnd] = "TagEnd"
	lex.TypeNames[ItemBool] = "Bool"
	lex.TypeNames[ItemField] = "Field"
	lex.TypeNames[ItemSet] = "Set"
	lex.TypeNames[ItemPlus] = "Plus"
	lex.TypeNames[ItemMinus] = "Minus"
	lex.TypeNames[ItemAsterisk] = "Asterisk"
	lex.TypeNames[ItemSlash] = "Slash"
	lex.TypeNames[ItemVerticalSlash] = "VerticalSlash"
	lex.TypeNames[ItemAssign] = "Assign"
	lex.TypeNames[ItemOpenSquareBracket] = "OpenSquareBracket"
	lex.TypeNames[ItemCloseSquareBracket] = "CloseSquareBracket"
	lex.TypeNames[ItemWrapper] = "Wrapper"
	lex.TypeNames[ItemComma] = "Comma"
	lex.TypeNames[ItemOpenParen] = "OpenParen"
	lex.TypeNames[ItemCloseParen] = "CloseParen"
	lex.TypeNames[ItemPeriod] = "Period"
	lex.TypeNames[ItemKeyword] = "Keyword"
	lex.TypeNames[ItemGet] = "GET"
	lex.TypeNames[ItemMacro] = "Macro"
	lex.TypeNames[ItemBlock] = "Block"
	lex.TypeNames[ItemDoubleQuotedString] = "DoubleQuotedString"
	lex.TypeNames[ItemSingleQuotedString] = "SingleQuotedString"
	lex.TypeNames[ItemWith] = "With"
	lex.TypeNames[ItemForeach] = "Foreach"
	lex.TypeNames[ItemWhile] = "While"
	lex.TypeNames[ItemIn] = "In"
	lex.TypeNames[ItemInclude] = "Include"
	lex.TypeNames[ItemIf] = "If"
	lex.TypeNames[ItemElse] = "Else"
	lex.TypeNames[ItemElseIf] = "ElseIf"
	lex.TypeNames[ItemUnless] = "Unless"
	lex.TypeNames[ItemSwitch] = "Switch"
	lex.TypeNames[ItemCase] = "Case"
	lex.TypeNames[ItemDefault] = "Default"
	lex.TypeNames[ItemCall] = "Call"
	lex.TypeNames[ItemOperator] = "Operator (INTERNAL)"
	lex.TypeNames[ItemRange] = "Range"
	lex.TypeNames[ItemEquals] = "Equals"
	lex.TypeNames[ItemNotEquals] = "NotEquals"
	lex.TypeNames[ItemCmp] = "Cmp"
	lex.TypeNames[ItemGT] = "GreaterThan"
	lex.TypeNames[ItemLT] = "LessThan"
	lex.TypeNames[ItemLE] = "LessThanEquals"
	lex.TypeNames[ItemGE] = "GreterThanEquals"
	lex.TypeNames[ItemShiftLeft] = "ShiftLeft"
	lex.TypeNames[ItemShiftRight] = "ShiftRight"
	lex.TypeNames[ItemAssignAdd] = "AssignAdd"
	lex.TypeNames[ItemAssignSub] = "AssignSub"
	lex.TypeNames[ItemAssignMul] = "AssignMul"
	lex.TypeNames[ItemAssignDiv] = "AssignDiv"
	lex.TypeNames[ItemAssignMod] = "AssignMod"
	lex.TypeNames[ItemAnd] = "And"
	lex.TypeNames[ItemOr] = "Or"
	lex.TypeNames[ItemFatComma] = "FatComma"
	lex.TypeNames[ItemIncr] = "Incr"
	lex.TypeNames[ItemDecr] = "Decr"
	lex.TypeNames[ItemMod] = "Mod"
	lex.TypeNames[ItemEnd] = "End"
}

func (l *Lexer) SetTagStart(s string) {
	l.tagStart = s
}

func (l *Lexer) SetTagEnd(s string) {
	l.tagEnd = s
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func isChar(r rune) bool {
	return r <= unicode.MaxASCII && unicode.IsPrint(r)
}

func isNumeric(r rune) bool {
	return '0' <= r && r <= '9'
}

func NewStringLexer(template string, ss *LexSymbolSet) *Lexer {
	l := &Lexer{
		nil,
		"",
		"",
		ss,
	}
	l.Lexer = lex.NewStringLexer(template, l.lexRawString)
	return l
}

func NewReaderLexer(rdr io.Reader, ss *LexSymbolSet) *Lexer {
	l := &Lexer{
		nil,
		"",
		"",
		ss,
	}
	l.Lexer = lex.NewReaderLexer(rdr, l.lexRawString)
	return l
}

func (sl *Lexer) lexRawString(l lex.Lexer) lex.LexFn {
	for {
		if sl.PeekString(sl.tagStart) {
			if len(l.BufferString()) > 0 {
				sl.Emit(ItemRawString)
			}
			return sl.lexTagStart
		}
		if sl.Next() == lex.EOF {
			break
		}
	}

	if len(sl.BufferString()) > 0 {
		sl.Emit(ItemRawString)
	}
	sl.Emit(ItemEOF)
	return nil
}

func (sl *Lexer) lexSpace(l lex.Lexer) lex.LexFn {
	guard := lex.Mark("lexSpace")
	defer guard()

	count := 0
	for {
		r := l.Peek()
		if !isSpace(r) {
			break
		}
		count++
		l.Next()
	}

	if count > 0 {
		l.Emit(ItemSpace)
	}
	return sl.lexInsideTag
}

func (sl *Lexer) lexTagStart(l lex.Lexer) lex.LexFn {
	if !sl.AcceptString(sl.tagStart) {
		sl.EmitErrorf("Expected tag start (%s)", sl.tagStart)
	}
	sl.Emit(ItemTagStart)
	return sl.lexInsideTag
}

func (sl *Lexer) lexTagEnd(l lex.Lexer) lex.LexFn {
	if !sl.AcceptString(sl.tagEnd) {
		sl.EmitErrorf("Expected tag end (%s)", sl.tagEnd)
	}
	sl.Emit(ItemTagEnd)
	return sl.lexRawString
}

func (sl *Lexer) lexIdentifier(l lex.Lexer) lex.LexFn {
Loop:
	for {
		switch r := sl.Next(); {
		case isAlphaNumeric(r):
		default:
			sl.Backup()
			word := sl.BufferString()
			if !sl.atTerminator() {
				return sl.EmitErrorf("bad character %#U", r)
			}

			if sym := sl.symbols.Get(word); sym.Type > ItemKeyword {
				sl.Emit(sym.Type)
			} else {
				switch {
				case word[0] == '.':
					sl.Emit(ItemField)
				case word == "true", word == "false":
					sl.Emit(ItemBool)
				default:
					sl.Emit(ItemIdentifier)
				}
			}
			break Loop
		}
	}
	return sl.lexInsideTag
}

func (l *Lexer) atTerminator() bool {
	r := l.Peek()
	if isSpace(r) || isEndOfLine(r) {
		return true
	}
	switch r {
	case lex.EOF, '.', ',', '|', ':', ')', '(', '[', ']':
		return true
	}
	// Does r start the delimiter? This can be ambiguous (with delim=="//", $x/2 will
	// succeed but should fail) but only in extremely rare cases caused by willfully
	// bad choice of delimiter.
	if rd, _ := utf8.DecodeRuneInString(l.tagEnd); rd == r {
		return true
	}
	return false
}

func (sl *Lexer) lexRange(l lex.Lexer) lex.LexFn {
	for i := 0; i < 2; i++ {
		if sl.Peek() != '.' {
			return sl.EmitErrorf("bad range syntax: %q", sl.BufferString())
		}
		sl.Next()
	}
	sl.Emit(ItemRange)

	if isNumeric(sl.Peek()) {
		return sl.lexInteger
	}
	return sl.lexIdentifier
}

func (sl *Lexer) lexInteger(l lex.Lexer) lex.LexFn {
	if sl.scanInteger() {
		sl.Emit(ItemNumber)
	} else {
		return sl.EmitErrorf("bad integer syntax: %q", sl.BufferString())
	}
	return sl.lexInsideTag
}

func (sl *Lexer) lexNumber(l lex.Lexer) lex.LexFn {
	if !sl.scanNumber() {
		return sl.EmitErrorf("bad number syntax: %q", sl.BufferString())
	}

	/* XXX Remove complex number support
	   if sign := sl.Peek(); sign == '+' || sign == '-' {
	     // Complex: 1+2i. No spaces, must end in 'i'.
	     if !sl.scanNumber() || sl.PrevByte() != 'i' {
	       return sl.EmitErrorf("bad number syntax: %q", sl.BufferString())
	     }
	     sl.Emit(ItemComplex)
	   } else
	*/
	if dot := sl.Peek(); dot == '.' {
		sl.Emit(ItemNumber)
		return sl.lexRange
	}
	sl.Emit(ItemNumber)
	return sl.lexInsideTag
}

func (l *Lexer) scanInteger() bool {
	l.AcceptAny("+-")
	digits := "0123456789"
	ret := l.AcceptRun(digits)
	//  l.backup()
	return ret
}

func (l *Lexer) scanNumber() bool {
	// Optional leading sign.
	l.AcceptAny("+-")
	// Is it hex?
	digits := "0123456789"
	if l.AcceptAny("0") && l.AcceptAny("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.AcceptRun(digits)
	if l.AcceptString(".") {
		if !l.AcceptRun(digits) {
			l.Backup()
		}
		return true
	}
	if l.AcceptAny("eE") {
		l.AcceptAny("+-")
		l.AcceptRun("0123456789")
	}
	// Is it imaginary?
	l.AcceptString("i")
	// Next thing mustn't be alphanumeric.
	if isAlphaNumeric(l.Peek()) {
		l.Next()
		return false
	}
	return true
}

func (sl *Lexer) lexComment(l lex.Lexer) lex.LexFn {
	for {
		if sl.PeekString(sl.tagEnd) {
			sl.Emit(ItemComment)
			return sl.lexTagEnd
		}
		if isEndOfLine(sl.Next()) {
			sl.Emit(ItemComment)
			return sl.lexTagEnd
		}
	}
}

func (sl *Lexer) lexQuotedString(l lex.Lexer, quote rune, t lex.ItemType) lex.LexFn {
	for {
		if sl.PeekString(sl.tagEnd) {
			return sl.EmitErrorf("unexpected end of quoted string")
		}

		r := sl.Next()
		switch r {
		case quote:
			sl.Emit(t)
			return sl.lexInsideTag
		case lex.EOF:
			return sl.EmitErrorf("unexpected end of quoted string")
		}
	}
}

func (sl *Lexer) lexDoubleQuotedString(l lex.Lexer) lex.LexFn {
	return sl.lexQuotedString(l, '"', ItemDoubleQuotedString)
}

func (sl *Lexer) lexSingleQuotedString(l lex.Lexer) lex.LexFn {
	return sl.lexQuotedString(l, '\'', ItemSingleQuotedString)
}

func (l *Lexer) getSortedSymbols() LexSymbolList {
	return l.symbols.GetSortedList()
}

func (sl *Lexer) lexInsideTag(l lex.Lexer) lex.LexFn {
	guard := lex.Mark("lexInsideTag")
	defer guard()

	if sl.PeekString(sl.tagEnd) {
		return sl.lexTagEnd
	}

	// Find registered symbols
	for _, sym := range sl.getSortedSymbols() {
		if sl.AcceptString(sym.Name) {
			sl.Emit(sym.Type)
			return sl.lexInsideTag
		}
	}

	r := sl.Next()
	lex.Trace("r = '%q'\n", r)
	switch {
	case r == lex.EOF:
		return sl.EmitErrorf("unclosed tag")
	case r == '#':
		return sl.lexComment
	case isSpace(r):
		sl.Backup()
		return sl.lexSpace
	case isNumeric(r):
		sl.Backup()
		return sl.lexNumber
	case r == '"':
		return sl.lexDoubleQuotedString
	case r == '\'':
		return sl.lexSingleQuotedString
	case isAlphaNumeric(r):
		sl.Backup()
		return sl.lexIdentifier
	default:
		return sl.EmitErrorf("unrecognized character in tag: %#U", r)
	}
}
