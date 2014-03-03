package vm

import (
  "html"
  "reflect"
  "unicode"
  "unicode/utf8"
)

const (
  TXOP_noop OpType = iota
  TXOP_nil
  TXOP_move_to_sb
  TXOP_move_from_sb
  TXOP_literal
  TXOP_fetch_s
  TXOP_fetch_field_s
  TXOP_print_raw
  TXOP_save_to_lvar
  TXOP_load_lvar
  TXOP_add
  TXOP_sub
  TXOP_mul
  TXOP_div
  TXOP_and
  TXOP_goto
  TXOP_for_start
  TXOP_for_iter
  TXOP_html_escape
  TXOP_uri_escape
  TXOP_end
  TXOP_max
)

var opnames    []string    = make([]string, TXOP_max)
var ophandlers []OpHandler = make([]OpHandler, TXOP_max)
var execcodes  []*ExecCode = make([]*ExecCode, TXOP_max)
func init () {
  for i := TXOP_noop; i < TXOP_max; i++ {
    var h OpHandler
    n := "Unknown"
    switch i {
    case TXOP_noop:
      h = txNoop
      n = "noop"
    case TXOP_end:
      h = txEnd
      n = "end"
    case TXOP_move_to_sb:
      h = txMoveToSb
      n = "move_to_sb"
    case TXOP_move_from_sb:
      h = txMoveFromSb
      n = "move_from_sb"
    case TXOP_print_raw:
      h = txPrintRaw
      n = "print_raw"
    case TXOP_literal:
      h = txLiteral
      n = "literal"
    case TXOP_fetch_s:
      h = txFetchSymbol
      n = "fetch_s"
    case TXOP_fetch_field_s:
      h = txFetchField
      n = "fetch_field_s"
    case TXOP_save_to_lvar:
      h = txSaveToLvar
      n = "save_to_lvar"
    case TXOP_load_lvar:
      h = txLoadLvar
      n = "load_lvar"
    case TXOP_nil:
      h = txNil
      n = "nil"
    case TXOP_add:
      h = txAdd
      n = "add"
    case TXOP_sub:
      h = txSub
      n = "sub"
    case TXOP_mul:
      h = txMul
      n = "mul"
    case TXOP_div:
      h = txDiv
      n = "div"
    case TXOP_and:
      h = txAnd
      n = "and"
    case TXOP_goto:
      h = txGoto
      n = "goto"
    case TXOP_for_start:
      h = txForStart
      n = "for_start"
    case TXOP_for_iter:
      h = txForIter
      n = "for_iter"
    case TXOP_html_escape:
      h = txHtmlEscape
      n = "html_escape"
    case TXOP_uri_escape:
      h = txUriEscape
      n = "uri_escape"
    default:
      panic("No such optype")
    }
    ophandlers[i] = h
    execcodes[i]  = &ExecCode { OpType(i), h}
    opnames[i]    = n
  }
}

func optypeToExecCode(o OpType) *ExecCode {
  return execcodes[o]
}

func optypeToHandler(o OpType) OpHandler {
  return ophandlers[o]
}

func txEnd(st *State) {}

func txNil(st *State) {
  st.sa = nil
  st.Advance()
}

func txNoop(st *State) {
  st.Advance()
}

func txMoveToSb(st *State) {
  st.sb = st.sa
  st.Advance()
}

func txMoveFromSb(st *State) {
  st.sa = st.sb
  st.Advance()
}

func txLiteral(st *State) {
  st.sa = st.CurrentOp().Arg()
  st.Advance()
}

func txFetchSymbol(st *State) {
  key   := st.CurrentOp().Arg()
  vars  := st.Vars()
  if v, ok := vars.Get(key); ok {
    st.sa = v
  } else {
    st.sa = nil
  }
  st.Advance()
}

func txFetchField(st *State) {
  container := st.sa
  if container == nil {
    // XXX ? no op?
    st.sa = nil
  } else {
    t := reflect.TypeOf(container)
    var v reflect.Value
    switch t.Kind() {
    case reflect.Ptr, reflect.Struct:
      v = reflect.ValueOf(container)
    default:
      v = reflect.ValueOf(&container).Elem()
    }
    name := st.CurrentOp().ArgString()
    r, size := utf8.DecodeRuneInString(name)
    name = string(unicode.ToUpper(r)) + name[size:]
    f := v.FieldByName(name)
    st.sa = f.Interface()
  }
  st.Advance()
}

func txPrintRaw(st *State) {
  // mark_raw handling
  arg := st.sa
  if arg == nil {
    st.Warnf("Use of nil to print\n")
  } else {
    v := interfaceToString(arg)
    st.AppendOutput([]byte(v))
  }
  st.Advance()
}

func txSaveToLvar(st *State) {
  idx := st.CurrentOp().ArgInt()
  st.CurrentFrame().SetLvar(idx, st.sa)
  st.Advance()
}

func txLoadLvar(st *State) {
  idx := st.CurrentOp().ArgInt()
  st.sa = st.CurrentFrame().GetLvar(idx)
  st.Advance()
}

func txAdd(st *State) {
  leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
  switch leftV.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    st.sa = leftV.Int() + rightV.Int()
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    st.sa = leftV.Uint() + rightV.Uint()
  case reflect.Float32, reflect.Float64:
    st.sa = leftV.Float() + rightV.Float()
  }

  // XXX: set to targ?
  st.Advance()
}

func txSub(st *State) {
  leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
  switch leftV.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    st.sa = leftV.Int() - rightV.Int()
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    st.sa = leftV.Uint() - rightV.Uint()
  case reflect.Float32, reflect.Float64:
    st.sa = leftV.Float() - rightV.Float()
  }

  // XXX: set to targ?
  st.Advance()
}

func txMul(st *State) {
  leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
  switch leftV.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    st.sa = leftV.Int() * rightV.Int()
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    st.sa = leftV.Uint() * rightV.Uint()
  case reflect.Float32, reflect.Float64:
    st.sa = leftV.Float() * rightV.Float()
  }

  // XXX: set to targ?
  st.Advance()
}

func txDiv(st *State) {
  leftV, rightV := alignTypesForArithmetic(st.sb, st.sa)
  switch leftV.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    // XXX This is a hack. We rely on interfaceToString() using FormatFloat(prec = -1)
    // to get rid of the fractional portions when printing
    typeF := reflect.TypeOf(0.1)
    st.sa = leftV.Convert(typeF).Float() / rightV.Convert(typeF).Float()
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    st.sa = leftV.Uint() / rightV.Uint()
  case reflect.Float32, reflect.Float64:
    st.sa = leftV.Float() / rightV.Float()
  }

  // XXX: set to targ?
  st.Advance()
}

func txAnd(st *State) {
  if interfaceToBool(st.sa) {
    st.Advance()
  } else {
    st.AdvanceBy(st.CurrentOp().ArgInt())
  }
}

func txGoto(st *State) {
  st.AdvanceBy(st.CurrentOp().ArgInt())
}

func txForStart(st *State) {
  id    := st.CurrentOp().ArgInt()
  slice := reflect.ValueOf(st.sa)

  switch slice.Kind() {
  case reflect.Array, reflect.Slice:
    // Normal case. nothing to do
  default:
    // Oh you silly goose. You didn't give me a slice.
    // Use a dummy array
    slice = reflect.ValueOf([]struct{}{})
  }

  cf := st.CurrentFrame()
  cf.SetLvar(id    , nil)   // item
  cf.SetLvar(id + 1, -1)    // index
  cf.SetLvar(id + 2, slice) // slice (Value)

  st.Advance()
}

func txForIter(st *State) {
  id    := st.sa.(int)
  cf    := st.CurrentFrame()
  index := cf.GetLvar(id + 1).(int)
  slice := cf.GetLvar(id + 2).(reflect.Value)

  index++
  cf.SetLvar(id + 1, index)
  if slice.Len() > index {
    cf.SetLvar(id, slice.Index(index).Interface())
    st.Advance()
    return
  }

  // loop done
  st.AdvanceBy(st.CurrentOp().ArgInt())
}

func txUriEscape(st *State) {
  v := interfaceToString(st.sa)
  st.sa = escapeUriString(v)
  st.Advance()
}

func txHtmlEscape(st *State) {
  v := interfaceToString(st.sa)
  st.sa = html.EscapeString(v)
  st.Advance()
}

