package parser

import (
  "sort"
)

var DefaultSymbolSet = NewLexSymbolSet()

func init() {
  DefaultSymbolSet.Set("=",   ItemAssign,             1.0)
  DefaultSymbolSet.Set("==",  ItemEquals,             0.0)
  DefaultSymbolSet.Set("!=",  ItemNotEquals,          0.0)
  DefaultSymbolSet.Set("(",   ItemOpenParen,          0.0)
  DefaultSymbolSet.Set(")",   ItemCloseParen,         0.0)
  DefaultSymbolSet.Set("[",   ItemOpenSquareBracket,  0.0)
  DefaultSymbolSet.Set("]",   ItemCloseSquareBracket, 0.0)
  DefaultSymbolSet.Set(".",   ItemPeriod,             0.0)
  DefaultSymbolSet.Set(",",   ItemComma,              0.0)
  DefaultSymbolSet.Set("|",   ItemVerticalSlash,      0.0)
  DefaultSymbolSet.Set(">",   ItemGT,                 0.0)
  DefaultSymbolSet.Set("<",   ItemLT,                 0.0)
  DefaultSymbolSet.Set("+",   ItemPlus)
  DefaultSymbolSet.Set("-",   ItemMinus)
  DefaultSymbolSet.Set("*",   ItemAsterisk)
  DefaultSymbolSet.Set("/",   ItemSlash)
}

// LexSymbol holds the pre-defined symbols to be lexed
type LexSymbol struct {
  Name string
  Type LexItemType
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
  return s.list[i].Priority < s.list[j].Priority
}

// Swap swaps the elements at i and j
func (s LexSymbolSorter) Swap(i, j int) {
  s.list[i], s.list[j] = s.list[j], s.list[i]
}

// LexSymbolSet 
type LexSymbolSet struct {
  Map         map[string]LexSymbol
  SortedList  LexSymbolList
}

func NewLexSymbolSet() *LexSymbolSet {
  return &LexSymbolSet {
    make(map[string]LexSymbol),
    nil,
  }
}

func (l *LexSymbolSet) Copy() *LexSymbolSet {
  c := NewLexSymbolSet()
  for k, v := range l.Map {
    c.Map[k] = LexSymbol { v.Name, v.Type, v.Priority }
  }
  return c
}

func (l *LexSymbolSet) Count() int {
  return len(l.Map)
}

func (l *LexSymbolSet) Get(name string) LexSymbol {
  return l.Map[name]
}

func (l *LexSymbolSet) Set(name string, typ LexItemType, prio ...float32) {
  var x float32
  if len(prio) < 1 {
    x = 1.0
  } else {
    x = prio[0]
  }
  l.Map[name] = LexSymbol { name, typ, x }
  l.SortedList = nil // reset
}

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


