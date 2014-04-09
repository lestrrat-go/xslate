package parser

import (
  "github.com/lestrrat/go-lex"
  "sort"
)

// DefaultSymbolSet is the LexSymbolSet for symbols that are common to
// all syntax
var DefaultSymbolSet = NewLexSymbolSet()

func init() {
  DefaultSymbolSet.Set("==",  ItemEquals,             1.0)
  DefaultSymbolSet.Set("eq",  ItemEquals,             0.0)
  DefaultSymbolSet.Set("!=",  ItemNotEquals,          1.0)
  DefaultSymbolSet.Set("ne",  ItemNotEquals,          0.0)
  DefaultSymbolSet.Set("+=",  ItemAssignAdd,          1.0)
  DefaultSymbolSet.Set("-=",  ItemAssignSub,          1.0)
  DefaultSymbolSet.Set("*=",  ItemAssignMul,          1.0)
  DefaultSymbolSet.Set("/=",  ItemAssignDiv,          1.0)
  DefaultSymbolSet.Set("(",   ItemOpenParen,          0.0)
  DefaultSymbolSet.Set(")",   ItemCloseParen,         0.0)
  DefaultSymbolSet.Set("[",   ItemOpenSquareBracket,  0.0)
  DefaultSymbolSet.Set("]",   ItemCloseSquareBracket, 0.0)
  DefaultSymbolSet.Set("..",  ItemRange,              1.0)
  DefaultSymbolSet.Set(".",   ItemPeriod,             0.0)
  DefaultSymbolSet.Set(",",   ItemComma,              0.0)
  DefaultSymbolSet.Set("|",   ItemVerticalSlash,      0.0)
  DefaultSymbolSet.Set("=",   ItemAssign,             0.0)
  DefaultSymbolSet.Set(">",   ItemGT,                 0.0)
  DefaultSymbolSet.Set("<",   ItemLT,                 0.0)
  DefaultSymbolSet.Set("+",   ItemPlus,               0.0)
  DefaultSymbolSet.Set("-",   ItemMinus,              0.0)
  DefaultSymbolSet.Set("*",   ItemAsterisk,           0.0)
  DefaultSymbolSet.Set("/",   ItemSlash,              0.0)
}

// LexSymbol holds the pre-defined symbols to be lexed
type LexSymbol struct {
  Name string
  Type lex.ItemType
  Priority float32
}

// LexSymbolList a list of LexSymbols. Normally you do not need to use it.
// This is mainly only useful for sorting LexSymbols
type LexSymbolList []LexSymbol

// Sort returns a sorted list of LexSymbols, sorted by Priority
func (list LexSymbolList) Sort() LexSymbolList {
  sorter := LexSymbolSorter {
    list: list,
  }
  sort.Sort(sorter)
  return sorter.list
}

// LexSymbolSorter sorts a list of LexSymbols by priority
type LexSymbolSorter struct {
  list LexSymbolList
}

// Len returns the length of the list
func (s LexSymbolSorter) Len() int {
  return len(s.list)
}

// Less returns true if the i-th element's Priority is less than the j-th element
func (s LexSymbolSorter) Less(i, j int) bool {
  return s.list[i].Priority > s.list[j].Priority
}

// Swap swaps the elements at i and j
func (s LexSymbolSorter) Swap(i, j int) {
  s.list[i], s.list[j] = s.list[j], s.list[i]
}

// LexSymbolSet is the container for symbols.
type LexSymbolSet struct {
  Map         map[string]LexSymbol
  SortedList  LexSymbolList
}

// NewLexSymbolSet creates a new LexSymbolSet
func NewLexSymbolSet() *LexSymbolSet {
  return &LexSymbolSet {
    make(map[string]LexSymbol),
    nil,
  }
}

// Copy creates a new copy of the given LexSymbolSet
func (l *LexSymbolSet) Copy() *LexSymbolSet {
  c := NewLexSymbolSet()
  for k, v := range l.Map {
    c.Map[k] = LexSymbol { v.Name, v.Type, v.Priority }
  }
  return c
}

// Count returns the number of symbols registered
func (l *LexSymbolSet) Count() int {
  return len(l.Map)
}

// Get returns the LexSymbol associated with `name`
func (l *LexSymbolSet) Get(name string) LexSymbol {
  return l.Map[name]
}

// Set creates and sets a new LexItem to `name`
func (l *LexSymbolSet) Set(name string, typ lex.ItemType, prio ...float32) {
  var x float32
  if len(prio) < 1 {
    x = 1.0
  } else {
    x = prio[0]
  }
  l.Map[name] = LexSymbol { name, typ, x }
  l.SortedList = nil // reset
}

// GetSortedList returns the lsit of LexSymbols in order that they should
// be searched for in the tempalte
func (l *LexSymbolSet) GetSortedList() LexSymbolList {
  // Because symbols are parsed automatically in a loop, we need to make
  // sure that we search starting with the longest term (e.g., "INCLUDE"
  // must come before "IN")
  // However, simply sorting the symbols using alphabetical sort then
  // max-length forces us to make more comparisons than necessary.
  // To get the best of both world, we allow passing a floating point
  // "priority" parameter to sort the symbols
  if l.SortedList != nil {
    return l.SortedList
  }

  num := len(l.Map)
  list := make(LexSymbolList, num)
  i := 0
  for _, v := range l.Map {
    list[i] = v
    i++
  }
  l.SortedList = list.Sort()

  return l.SortedList
}


