package parser

/* 

Lexer for TTerse, based on http://golang.org/src/pkg/text/template/parse/lex.go

Anything up to a tagStart('[%') is treated as RawText, and therefore does not
need any real lexing.

Once tagStart is found, real lexing starts.

*/

import (
  "fmt"
  "strings"
  "unicode"
  "unicode/utf8"
)

type LexItemType int
type LexItem struct {
  typ LexItemType
  pos Pos
  val string
}
const eof = -1

func NewLexItem(t LexItemType, p Pos, v string) LexItem {
  return LexItem { t, p, v }
}

func (l LexItem) Copy() LexItem {
  return NewLexItem(l.typ, l.pos, l.val)
}

func (l LexItem) Type() LexItemType {
  return l.typ
}

func (l LexItem) Pos() Pos {
  return l.pos
}

func (l LexItem) Value() string {
  return l.val
}

const (
  ItemError       LexItemType = iota
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
  ItemOpenParen   // '('
  ItemCloseParen  // ')'
  ItemOpenSquareBracket   // '['
  ItemCloseSquareBracket  // ']'
  ItemPeriod      // '.'
  ItemKeyword     // Delimiter
  ItemGet         // GET
  ItemSet         // SET
  ItemMacro       // MACRO
  ItemBlock       // BLOCK
  ItemForeach     // FOREACH
  ItemIn          // IN
  ItemInclude     // INCLUDE
  ItemWith        // WITH
  ItemIf          // IF
  ItemElse        // ELSE
  ItemElseIf      // ELSIF
  ItemUnless      // UNLESS
  ItemSwitch      // SWITCH
  ItemCase        // CASE
  ItemWrapper     // WRAPPER
  ItemDefault     // DEFAULT
  ItemCall        // CALL
  ItemEnd         // END
  ItemOperator    // Delimiter
  ItemRange       // ..
  ItemEquals      // ==
  ItemNotEquals   // !=
  ItemGT          // >
  ItemLT          // <
  ItemCmp         // <=>
  ItemLE          // <=
  ItemGE          // >=
  ItemShiftLeft   // <<
  ItemShiftRight  // >>
  ItemAssignAdd   // +=
  ItemAssignSub   // -=
  ItemAssignMul   // *=
  ItemAssignDiv   // /=
  ItemAssignMod   // %=
  ItemAnd         // &&
  ItemOr          // ||
  ItemFatComma    // =>
  ItemIncr        // ++
  ItemDecr        // --
  ItemPlus
  ItemMinus
  ItemAsterisk
  ItemSlash
  ItemVerticalSlash
  ItemMod
  ItemAssign      // =

  DefaultItemTypeMax
)

type Lexer struct {
//  name  string
  input string
  inputLength int
  start int
  pos   int
  width int
  tagStart string
  tagEnd   string
  symbols       *LexSymbolSet
  items chan LexItem
}

type LexRunner interface {
  Run()
  NextItem() LexItem
}

type stateFn func(*Lexer) stateFn

func (l *Lexer) SetInput(s string) {
  l.input = s
  l.inputLength = len(s)
}

func (l *Lexer) SetTagStart(s string) {
  l.tagStart = s
}

func (l *Lexer) SetTagEnd(s string) {
  l.tagEnd = s
}

func (i LexItemType) String() string {
  var name string
  switch i {
  case ItemError:
    name = "Error"
  case ItemRawString:
    name = "RawString"
  case ItemEOF:
    name = "EOF"
  case ItemComment:
    name = "Comment"
  case ItemComplex:
    name = "Complex" // may not need this
  case ItemChar:
    name = "Char"
  case ItemSpace:
    name = "Space"
  case ItemNumber:
    name = "Number"
  case ItemSymbol:
    name = "Symbol"
  case ItemIdentifier:
    name = "Identifier"
  case ItemTagStart:
    name = "TagStart"
  case ItemTagEnd:
    name = "TagEnd"
  case ItemBool:
    name = "Bool"
  case ItemField:
    name = "Field"
  case ItemSet:
    name = "Set"
  case ItemPlus:
    name = "Plus"
  case ItemMinus:
    name = "Minus"
  case ItemAsterisk:
    name = "Asterisk"
  case ItemSlash:
    name = "Slash"
  case ItemVerticalSlash:
    name = "VerticalSlash"
  case ItemAssign:
    name = "Assign"
  case ItemOpenSquareBracket:
    name = "OpenSquareBracket"
  case ItemCloseSquareBracket:
    name = "CloseSquareBracket"
  case ItemWrapper:
    name = "Wrapper"
  case ItemComma:
    name = "Comma"
  case ItemOpenParen:
    name = "OpenParen"
  case ItemCloseParen:
    name = "CloseParen"
  case ItemPeriod:
    name = "Period"
  case ItemKeyword:
    name = "Keyword"
  case ItemGet:
    name = "GET"
  case ItemMacro:
    name = "Macro"
  case ItemBlock:
    name = "Block"
  case ItemDoubleQuotedString:
    name = "DoubleQuotedString"
  case ItemSingleQuotedString:
    name = "SingleQuotedString"
  case ItemWith:
    name = "With"
  case ItemForeach:
    name = "Foreach"
  case ItemIn:
    name = "In"
  case ItemInclude:
    name = "Include"
  case ItemIf:
    name = "If"
  case ItemElse:
    name = "Else"
  case ItemElseIf:
    name = "ElseIf"
  case ItemUnless:
    name = "Unless"
  case ItemSwitch:
    name = "Switch"
  case ItemCase:
    name = "Case"
  case ItemDefault:
    name = "Default"
  case ItemCall:
    name = "Call"
  case ItemOperator:
    name = "Operator (INTERNAL)"
  case ItemRange:
    name = "Range"
  case ItemEquals:
    name = "Equals"
  case ItemNotEquals:
    name = "NotEquals"
  case ItemCmp:
    name = "Cmp"
  case ItemGT:
    name = "GreaterThan"
  case ItemLT:
    name = "LessThan"
  case ItemLE:
    name = "LessThanEquals"
  case ItemGE:
    name = "GreterThanEquals"
  case ItemShiftLeft:
    name = "ShiftLeft"
  case ItemShiftRight:
    name = "ShiftRight"
  case ItemAssignAdd:
    name = "AssignAdd"
  case ItemAssignSub:
    name = "AssignSub"
  case ItemAssignMul:
    name = "AssignMul"
  case ItemAssignDiv:
    name = "AssignDiv"
  case ItemAssignMod:
    name = "AssignMod"
  case ItemAnd:
    name = "And"
  case ItemOr:
    name = "Or"
  case ItemFatComma:
    name = "FatComma"
  case ItemIncr:
    name = "Incr"
  case ItemDecr:
    name = "Decr"
  case ItemMod:
    name = "Mod"
  case ItemEnd:
    name = "End"
  default:
    name = fmt.Sprintf("Unknown Item (%d)", i)
  }
  return name
}

func (l LexItem) String() string {
  return fmt.Sprintf("%s (%s)", l.typ, l.val)
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

func NewLexer(ss *LexSymbolSet) *Lexer {
  l := &Lexer {
    items: make(chan LexItem, 1),
    symbols: ss,
  }
  return l
}

func lexRawString(l *Lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], l.tagStart) {
      if l.pos > l.start {
        l.Emit(ItemRawString)
      }
      return lexTagStart
    }
    if l.next() == eof {
      break
    }
  }

  if l.pos > l.start {
    l.Emit(ItemRawString)
  }

  l.Emit(ItemEOF)
  return nil
}

func lexSpace(l *Lexer) stateFn {
  for isSpace(l.peek()) {
    l.next()
  }
  l.Emit(ItemSpace)
  return lexInsideTag
}

func lexTagStart(l *Lexer) stateFn {
  l.pos += len(l.tagStart)
  l.Emit(ItemTagStart)
  return lexInsideTag
}

func lexTagEnd(l *Lexer) stateFn {
  l.pos += len(l.tagEnd)
  l.Emit(ItemTagEnd)
  return lexRawString
}

func lexIdentifier(l *Lexer) stateFn {
Loop:
  for {
    switch r := l.next(); {
    case isAlphaNumeric(r):
    default:
      l.backup()
      word := l.input[l.start:l.pos]
      if !l.atTerminator() {
        return l.errorf("bad character %#U", r)
      }

      if sym := l.symbols.Get(word); sym.Type > ItemKeyword {
        l.Emit(sym.Type)
      } else {
        switch {
          case word[0] == '.':
            l.Emit(ItemField)
          case word == "true", word == "false":
            l.Emit(ItemBool)
          default:
            l.Emit(ItemIdentifier)
        }
      }
      break Loop
    }
  }
  return lexInsideTag
}

func (l *Lexer) atTerminator() bool {
  r := l.peek()
  if isSpace(r) || isEndOfLine(r) {
    return true
  }
  switch r {
  case eof, '.', ',', '|', ':', ')', '(', '[':
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

func lexRange(l *Lexer) stateFn {
  for i := 0; i < 2; i++ {
    if l.peek() != '.' {
      return l.errorf("bad range syntax: %q", l.input[l.start:l.pos])
    }
    l.next()
  }
  l.Emit(ItemRange)

  return lexInteger
}
func lexInteger(l *Lexer) stateFn {
  if l.scanInteger() {
    l.Emit(ItemNumber)
  } else {
    l.errorf("bad integer syntax: %q", l.input[l.start:l.pos])
  }
  return lexInsideTag
}

func lexNumber(l *Lexer) stateFn {
  if !l.scanNumber() {
    return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
  }
  if sign := l.peek(); sign == '+' || sign == '-' {
    // Complex: 1+2i. No spaces, must end in 'i'.
    if !l.scanNumber() || l.input[l.pos-1] != 'i' {
      return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
    }
    l.Emit(ItemComplex)
  } else if dot := l.peek(); dot == '.' {
    l.Emit(ItemNumber)
    return lexRange
  } else {
    l.Emit(ItemNumber)
  }
  return lexInsideTag
}

func (l *Lexer) scanInteger() bool {
  l.accept("+-")
  digits := "0123456789"
  ret := l.acceptRun(digits)
//  l.backup()
  return ret
}

func (l *Lexer) scanNumber() bool {
  // Optional leading sign.
  l.accept("+-")
  // Is it hex?
  digits := "0123456789"
  if l.accept("0") && l.accept("xX") {
    digits = "0123456789abcdefABCDEF"
  }
  l.acceptRun(digits)
  if l.accept(".") {
    if ! l.acceptRun(digits) {
      l.backup()
    }
    return true
  }
  if l.accept("eE") {
    l.accept("+-")
    l.acceptRun("0123456789")
  }
  // Is it imaginary?
  l.accept("i")
  // Next thing mustn't be alphanumeric.
  if isAlphaNumeric(l.peek()) {
    l.next()
    return false
  }
  return true
}

func lexComment(l *Lexer) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], l.tagEnd) {
      l.Emit(ItemComment)
      return lexTagEnd
    }
    if isEndOfLine(l.next()) {
      l.Emit(ItemComment)
      return lexTagEnd
    }
  }
}

func lexQuotedString(l *Lexer, quote rune, t LexItemType) stateFn {
  for {
    if strings.HasPrefix(l.input[l.pos:], l.tagEnd) {
      return l.errorf("unexpected end of quoted string")
    }

    r := l.next()
    switch r {
    case quote:
      l.Emit(t)
      return lexInsideTag
    case eof:
      return l.errorf("unexpected end of quoted string")
    }
  }
}

func lexDoubleQuotedString(l *Lexer) stateFn {
  return lexQuotedString(l, '"', ItemDoubleQuotedString)
}

func lexSingleQuotedString(l *Lexer) stateFn {
  return lexQuotedString(l, '\'', ItemSingleQuotedString)
}

func (l *Lexer) getSortedSymbols() LexSymbolList {
  return l.symbols.GetSortedList()
}

func lexInsideTag(l *Lexer) stateFn {
  if strings.HasPrefix(l.input[l.pos:], l.tagEnd) {
    return lexTagEnd
  }

  // Find registered symbols
  for _, sym := range l.getSortedSymbols() {
    if strings.HasPrefix(l.input[l.pos:], sym.Name) {
      l.pos += len(sym.Name)
      l.Emit(sym.Type)
      return lexInsideTag
    }
  }

  switch r := l.next(); {
  case r == eof:
    return l.errorf("unclosed tag")
  case r == '#':
    return lexComment
  case isSpace(r):
    return lexSpace
  case isNumeric(r):
    l.backup()
    return lexNumber
  case r == '"':
    return lexDoubleQuotedString
  case r == '\'':
    return lexSingleQuotedString
  case isAlphaNumeric(r):
    l.backup()
    return lexIdentifier
  default:
    return l.errorf("unrecognized character in tag: %#U", r)
  }

  return lexInsideTag
}

func (l *Lexer) inputLen() int {
  if l.inputLength == -1 {
    l.inputLength = len(l.input)
  }
  return l.inputLength
}

func (l *Lexer) next() (r rune) {
  if l.pos >= l.inputLen() {
    l.width = 0
    return eof
  }

  r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
  l.pos += l.width
  return r
}

func (l *Lexer) peek() rune {
  r := l.next()
  l.backup()
  return r
}

func (l *Lexer) backup() {
  l.pos -= l.width
}

func (l *Lexer) accept(valid string) bool {
  if strings.IndexRune(valid, l.next()) >= 0 {
    return true
  }
  l.backup()
  return false
}

func (l *Lexer) acceptRun(valid string) bool {
  count := 0
  for strings.IndexRune(valid, l.next()) >= 0 {
    count++
  }
  l.backup()
  return count > 0
}

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
  l.items <-LexItem { ItemError, Pos(l.start), fmt.Sprintf(format, args...) }
  return nil
}

func (l *Lexer) Grab(t LexItemType) LexItem {
  return LexItem { t, Pos(l.start), l.input[l.start:l.pos] }
}

func (l *Lexer) Emit(t LexItemType) {
  l.items <-l.Grab(t)
  l.start = l.pos
}

func (l *Lexer) Run() {
  for state := lexRawString; state != nil; {
    state = state(l)
  }
  close(l.items)
}

func (l *Lexer) NextItem() LexItem {
  i := <-l.items
  return i
}
