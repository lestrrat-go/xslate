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
  pos int
  val string
}
const eof = -1

func NewLexItem(t LexItemType, p int, v string) LexItem {
  return LexItem { t, p, v }
}

func (l *LexItem) Type() LexItemType {
  return l.typ
}

func (l *LexItem) Pos() int {
  return l.pos
}

func (l *LexItem) Value() string {
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
  ItemOpenParen   // '('
  ItemCloseParen  // ')'
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
  ItemNotEqual    // !=
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
  ItemDiv
  ItemMul
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
  symbols map[string]LexItemType
  operators map[string]LexItemType
  items chan LexItem
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

func (l *Lexer) AddSymbol(s string, i LexItemType) {
  l.symbols[s] = i
}

func (l *Lexer) AddOperator(s string, i LexItemType) {
  l.operators[s] = i
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
  case ItemSpace:
    name = "Space"
  case ItemIdentifier:
    name = "Identifier"
  case ItemTagStart:
    name = "TagStart"
  case ItemTagEnd:
    name = "TagEnd"
  case ItemSet:
    name = "Set"
  case ItemPlus:
    name = "Plus"
  case ItemAssign:
    name = "Assign"
  default:
    name = fmt.Sprintf("Unknown(%d)", i)
  }
  return name
}

func (i LexItem) String() string {
  return fmt.Sprintf("%s (%s)", i.typ, i.val)
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

func NewLexer() *Lexer {
  l := &Lexer {
    items: make(chan LexItem, 1),
    symbols: make(map[string]LexItemType),
    operators: make(map[string]LexItemType),
  }
  l.AddSymbol("(", ItemOpenParen)
  l.AddSymbol(")", ItemCloseParen)
  l.AddSymbol(".", ItemPeriod)
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
      // absorb.
    default:
      l.backup()
      word := l.input[l.start:l.pos]
      if !l.atTerminator() {
        return l.errorf("bad character %#U", r)
      }
      switch {
      case l.symbols[word] > ItemKeyword:
        l.Emit(l.symbols[word])
      case word[0] == '.':
        l.Emit(ItemField)
      case word == "true", word == "false":
        l.Emit(ItemBool)
      default:
        l.Emit(ItemIdentifier)
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
  case eof, '.', ',', '|', ':', ')', '(':
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
  } else {
    l.Emit(ItemNumber)
  }
  return lexInsideTag
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
    l.acceptRun(digits)
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

func lexInsideTag(l *Lexer) stateFn {
  if strings.HasPrefix(l.input[l.pos:], l.tagEnd) {
    return lexTagEnd
  }

  // Find registered symbols
  for v, k := range l.symbols {
    if strings.HasPrefix(l.input[l.pos:], v) {
      l.pos += len(v)
      l.Emit(k)
      return lexInsideTag
    }
  }

  // Find registered operators
  for v, k := range l.operators {
    if strings.HasPrefix(l.input[l.pos:], v) {
      l.pos += len(v)
      l.Emit(k)
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

func (l *Lexer) acceptRun(valid string) {
  for strings.IndexRune(valid, l.next()) >= 0 {
  }
  l.backup()
}

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
  l.items <-LexItem { ItemError, l.start, fmt.Sprintf(format, args...) }
  return nil
}

func (l *Lexer) Emit(t LexItemType) {
  l.items <-LexItem { t, l.start, l.input[l.start:l.pos] }
  l.start = l.pos
}

func (l *Lexer) Run() {
  for state := lexRawString; state != nil; {
    state = state(l)
  }
  close(l.items)
}

func (l *Lexer) NextItem() LexItem {
  return <-l.items
}
