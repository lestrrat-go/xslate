package vm

import (
  "bytes"
  "encoding/binary"
  "fmt"
  "reflect"
)

// OpType is an integer identifying the type of op code
type OpType int

// Type returns the ... OpType. This seems redundunt, but having this method
// allows us to embed OpType in Op and have the ability to call Typ()
// without having to re-declare it
func (o OpType) Type() OpType {
  return o
}

// String returns the textual representation of an OpType
func (o OpType) String() string {
  return opnames[o]
}

// OpHandler describes an op's actual code
type OpHandler func(*State)

// Op represents a single op. It has an OpType, OpHandler, and an optional
// parameter to be used
type Op struct {
  OpType
  OpHandler
  uArg interface {}
}

// NewOp creates a new Op
func NewOp(o OpType, args ...interface {}) *Op {
  h := optypeToHandler(o)
  var arg interface {}
  if len(args) > 0 {
    arg = args[0]
  } else {
    arg = nil
  }
  return &Op { o, h, arg }
}

// MarshalBinary is used to serialize an Op into a binary form. This
// is used to cache the ByteCode
func (o Op) MarshalBinary() ([]byte, error) {
  buf := &bytes.Buffer {}

  // Write the code/opcode
  if err := binary.Write(buf, binary.LittleEndian, int64(o.OpType)); err != nil {
    return nil, fmt.Errorf("error: Op.MarshalBinary failed: %s", err)
  }

  // If this has args, we need to encode the args
  tArg   := reflect.TypeOf(o.uArg)
  hasArg := tArg != nil
  if hasArg {
    binary.Write(buf, binary.LittleEndian, int64(1))
  } else {
    binary.Write(buf, binary.LittleEndian, int64(0))
  }

  if hasArg {
    switch tArg.Kind() {
    case reflect.Int:
      binary.Write(buf, binary.LittleEndian, int64(2))
      binary.Write(buf, binary.LittleEndian, int64(o.uArg.(int)))
    case reflect.Int64:
      binary.Write(buf, binary.LittleEndian, int64(2))
      binary.Write(buf, binary.LittleEndian, int64(o.uArg.(int64)))
    case reflect.Slice:
      if tArg.Elem().Kind() != reflect.Uint8 {
        panic("Slice of what?")
      }
      binary.Write(buf, binary.LittleEndian, int64(5))
      binary.Write(buf, binary.LittleEndian, int64(len(o.uArg.([]byte))))
      for _, v := range o.uArg.([]byte) {
        binary.Write(buf, binary.LittleEndian, v)
      }
    case reflect.String:
      binary.Write(buf, binary.LittleEndian, int64(6))
      binary.Write(buf, binary.LittleEndian, int64(len(o.uArg.(string))))
      for _, v := range []byte(o.uArg.(string)) {
        binary.Write(buf, binary.LittleEndian, v)
      }
    default:
      panic("Unknown type " + tArg.String())
    }
  }

  return buf.Bytes(), nil
}

// UnmarshalBinary is used to deserialize an Op from binary form.
func (o *Op) UnmarshalBinary(data []byte) error {
  buf := bytes.NewReader(data)

  var t int64
  if err := binary.Read(buf, binary.LittleEndian, &t); err != nil {
    return fmt.Errorf("error: Op.UnmarshalBinary optype check failed: %s", err)
  }

  o.OpType    = OpType(t)
  o.OpHandler = optypeToHandler(o.OpType)

  var hasArg int64
  if err := binary.Read(buf, binary.LittleEndian, &hasArg); err != nil {
    return fmt.Errorf("error: Op.UnmarshalBinary hasArg check failed: %s", err)
  }

  if hasArg == 0 {
    // No args
    return nil
  }

  var tArg int64
  if err := binary.Read(buf, binary.LittleEndian, &tArg); err != nil {
    return fmt.Errorf("error: Op.UnmarshalBinary arg type check failed: %s", err)
  }

  switch tArg {
  case 2:
    var i int64
    if err := binary.Read(buf, binary.LittleEndian, &i); err != nil {
      return err
    }
    o.uArg = i
  case 5:
    var l int64
    if err := binary.Read(buf, binary.LittleEndian, &l); err != nil {
      return err
    }

    b := make([]byte, l)
    for i := int64(0); i < l; i++ {
      if err := binary.Read(buf, binary.LittleEndian, &b[i]); err != nil {
        return err
      }
    }
    o.uArg = b
  default:
    panic(fmt.Sprintf("Unknown tArg: %d", tArg))
  }

  return nil
}

// Call executes the Op code in the context of given vm.State
func (o *Op) Call(st *State) {
  o.OpHandler(st)
}

// SetArg sets the argument to this Op
func (o *Op) SetArg(v interface {}) {
  o.uArg = v
}

// Arg returns the Op code's argument
func (o *Op) Arg() interface {} {
  return o.uArg
}

// ArgInt returns the integer representation of the argument
func (o *Op) ArgInt() int {
  v := interfaceToNumeric(o.uArg)
  return int(v.Int())
}

// ArgString returns the string representatin of the argument
func (o *Op) ArgString() string {
  return interfaceToString(o.uArg)
}

func (o *Op) String() string {
  // TODO: also print out register id's and stuff
  if o.uArg != nil {
    return fmt.Sprintf("Op[%s] (%q)", o.Type(), o.ArgString())
  }

  return fmt.Sprintf("Op[%s]", o.Type())
}


