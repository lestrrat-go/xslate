package vm

import (
  "bytes"
  "encoding/binary"
  "encoding/json"
  "errors"
  "fmt"
  "reflect"
)

type OpType int
type OpHandler func(*State)
type ExecCode struct {
  id   OpType
  code func(*State)
}
type Op struct {
  code  *ExecCode
  u_arg interface {}
}

func NewOp(o OpType, args ...interface {}) *Op {
  e := optypeToExecCode(o)
  var arg interface {}
  if len(args) > 0 {
    arg = args[0]
  } else {
    arg = nil
  }
  return &Op { e, arg }
}

func (o ExecCode) MarshalBinary() ([]byte, error) {
  buf := &bytes.Buffer {}
  if err := binary.Write(buf, binary.LittleEndian, int64(o.id)); err != nil {
    return nil, errors.New(fmt.Sprintf("ExecCode.MarshalBinary: %s", err))
  }
  return buf.Bytes(), nil
}

func (o *ExecCode) UnmarshalBinary(data []byte) error {
  buf := bytes.NewReader(data)
  var t OpType
  if err := binary.Read(buf, binary.LittleEndian, &t); err != nil {
    return err
  }
  o = optypeToExecCode(t)
  return nil
}

func (o Op) MarshalBinary() ([]byte, error) {
  buf := &bytes.Buffer {}

  // Write the code/opcode
  b, err := o.code.MarshalBinary()
  if err != nil {
    return nil, err
  }

  buf.Write(b)

  // If this has args, we need to encode the args
  tArg   := reflect.TypeOf(o.u_arg)
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
      binary.Write(buf, binary.LittleEndian, int64(o.u_arg.(int)))
    case reflect.Slice:
      if tArg.Elem().Kind() != reflect.Uint8 {
        panic("Slice of what?")
      }
      binary.Write(buf, binary.LittleEndian, int64(5))
      binary.Write(buf, binary.LittleEndian, int64(len(o.u_arg.([]byte))))
      for _, v := range o.u_arg.([]byte) {
        binary.Write(buf, binary.LittleEndian, v)
      }
    default:
      panic("Unknown type " + tArg.String())
    }
  }

  return buf.Bytes(), nil
}

func (o *Op) UnmarshalBinary(data []byte) error {
  buf := bytes.NewReader(data)

  var t int64
  if err := binary.Read(buf, binary.LittleEndian, &t); err != nil {
    return errors.New(fmt.Sprintf("Op.UnmarshalBinary: error during optype check: %s", err))
  }
  o.code = optypeToExecCode(OpType(t))

  var hasArg int64
  if err := binary.Read(buf, binary.LittleEndian, &hasArg); err != nil {
    return errors.New(fmt.Sprintf("Op.UnmarshalBinary: error during hasArg check: %s", err))
  }

  if hasArg == 0 {
    // No args
    return nil
  }

  var tArg int64 = 0
  if err := binary.Read(buf, binary.LittleEndian, &tArg); err != nil {
    return errors.New(fmt.Sprintf("Op.UnmarshalBinary: error during arg type check: %s", err))
  }

  switch tArg {
  case 2:
    var i int64
    if err := binary.Read(buf, binary.LittleEndian, &i); err != nil {
      return err
    }
    o.u_arg = i
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
    o.u_arg = b
  default:
    panic(fmt.Sprintf("Unknown tArg: %d", tArg))
  }

  return nil
}

func (o Op) MarshalJSON() ([]byte, error) {
  return json.Marshal(map[string]interface{}{
    "code": o.code.id.String(),
    "code_id": o.code.id,
    "u_arg": o.u_arg,
  })
}

func (o OpType) String() string {
  return opnames[o]
}

func (o *Op) Call(st *State) {
  o.code.code(st)
}
func (o *Op) OpType() OpType {
  return o.code.id
}
func (o *Op) Code() *ExecCode {
  return o.code
}

func (o *Op) SetArg(v interface {}) {
  o.u_arg = v
}

func (o *Op) Arg() interface {} {
  return o.u_arg
}

func (o *Op) ArgInt() int {
  v := reflect.ValueOf(o.Arg())
  return int(v.Int())
}

func (o *Op) ArgString() string {
  return interfaceToString(o.Arg())
}

func (o *Op) String() string {
  // TODO: also print out register id's and stuff
  if o.u_arg != nil {
    return fmt.Sprintf("Op[%s] (%s)", o.OpType(), o.ArgString())
  } else {
    return fmt.Sprintf("Op[%s]", o.OpType())
  }
}


