package vm

import (
  "encoding/json"
  "fmt"
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

func (o Op) MarshalJSON() ([]byte, error) {
  return json.Marshal(map[string]interface{}{
    "code": o.code.id.String(),
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

func (o *Op) Arg() interface {} {
  return o.u_arg
}

func (o *Op) ArgInt() int {
  return o.Arg().(int)
}

func (o *Op) ArgString() string {
  return interfaceToString(o.Arg())
}

func (o *Op) String() string {
  // TODO: also print out register id's and stuff
  return fmt.Sprintf("Op[%s]", o.OpType())
}


