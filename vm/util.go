package vm

import (
  "fmt"
  "reflect"
  "strconv"
)

var hexdigits []byte = []byte("0123456789ABCDEF")
func escapeUri(thing []byte) []byte {
  ret := make([]byte, 0, len(thing))
  for _, v := range thing {
    if !shouldEscapeUri(v) {
      ret = append(ret, v)
    } else {
      ret = append(ret, '%')
      ret = append(ret, hexdigits[v & 0xf0 >> 4])
      ret = append(ret, hexdigits[v & 0x0f])
    }
  }

  return ret
}

func escapeUriString(thing string) string {
  return string(escapeUri([]byte(thing)))
}

func shouldEscapeUri(v byte) bool {
  switch v {
  case 0x2D, 0x2E, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5A, 0x5F, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69:
    return false
  default:
    return true
  }
}

func convertNumeric(v interface{}) reflect.Value {
  t := reflect.TypeOf(v)
  switch t.Kind() {
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
    return reflect.ValueOf(v)
  default:
    return reflect.ValueOf(0)
  }
}

// Given possibly non-matched pair of things to perform arithmetic
// operations on, align their types so that the given operation
// can be performed correctly.
// e.g. given int, float, we align them to float, float
func alignTypesForArithmetic(left, right interface {}) (reflect.Value, reflect.Value) {
  leftV  := convertNumeric(left)
  rightV := convertNumeric(right)

  if leftV.Kind() == rightV.Kind() {
    return leftV, rightV
  }

  var alignTo reflect.Type
  if leftV.Kind() > rightV.Kind() {
    alignTo = leftV.Type()
  } else {
    alignTo = rightV.Type()
  }

  return leftV.Convert(alignTo), rightV.Convert(alignTo)
}

func interfaceToString(arg interface {}) string {
  t := reflect.TypeOf(arg)
  var v string
  switch t.Kind() {
  case reflect.String:
    v = string(reflect.ValueOf(arg).String())
  case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
    v = strconv.FormatInt(reflect.ValueOf(arg).Int(), 10)
  case reflect.Float32, reflect.Float64:
    v = strconv.FormatFloat(reflect.ValueOf(arg).Float(), 'f', -1, 64)
  case reflect.Bool:
    if reflect.ValueOf(arg).Bool() {
      v = "true"
    } else {
      v = "false"
    }
  default:
    v = fmt.Sprintf("%s", arg)
  }
  return v
}

func interfaceToBool(arg interface {}) bool {
  t := reflect.TypeOf(arg)
  if t.Kind() == reflect.Bool {
    return arg.(bool)
  }

  z := reflect.Zero(t)
  return reflect.DeepEqual(z, t)
}

