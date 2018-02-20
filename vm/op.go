package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/lestrrat-go/xslate/internal/rbpool"
	"github.com/lestrrat-go/xslate/node"
	"github.com/pkg/errors"
)

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

// NewOp creates a new Op.
func NewOp(o OpType, args ...interface{}) Op {
	h := optypeToHandler(o)

	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	return &op{
		OpType:    o,
		OpHandler: h,
		uArg:      arg,
	}
}

func (o op) Comment() string {
	return o.comment
}

func (o op) Handler() OpHandler {
	return o.OpHandler
}

// MarshalBinary is used to serialize an Op into a binary form. This
// is used to cache the ByteCode
func (o op) MarshalBinary() ([]byte, error) {
	buf := rbpool.Get()
	defer rbpool.Release(buf)

	// Write the code/opcode
	if err := binary.Write(buf, binary.LittleEndian, int64(o.OpType)); err != nil {
		return nil, errors.Wrap(err, "failed to marshal op to binary")
	}

	// If this has args, we need to encode the args
	tArg := reflect.TypeOf(o.uArg)
	hasArg := tArg != nil
	if hasArg {
		binary.Write(buf, binary.LittleEndian, int8(1))
	} else {
		binary.Write(buf, binary.LittleEndian, int8(0))
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

	v := o.comment
	hasComment := v != ""
	if hasComment {
		binary.Write(buf, binary.LittleEndian, int8(1))
		binary.Write(buf, binary.LittleEndian, v)
	} else {
		binary.Write(buf, binary.LittleEndian, int8(0))
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary is used to deserialize an Op from binary form.
func (o *op) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)

	var t int64
	if err := binary.Read(buf, binary.LittleEndian, &t); err != nil {
		return errors.Wrap(err, "optype check failed during UnmarshalBinary")
	}

	o.OpType = OpType(t)
	o.OpHandler = optypeToHandler(o.OpType)

	var hasArg int8
	if err := binary.Read(buf, binary.LittleEndian, &hasArg); err != nil {
		return errors.Wrap(err, "hasArg check failed during UnmarshalBinary")
	}

	if hasArg == 1 {
		var tArg int64
		if err := binary.Read(buf, binary.LittleEndian, &tArg); err != nil {
			return errors.Wrap(err, "failed to read argument from buffer during UnmarshalBinary")
		}

		switch tArg {
		case 2:
			var i int64
			if err := binary.Read(buf, binary.LittleEndian, &i); err != nil {
				return errors.Wrap(err, "failed to read integer argument during UnmarshalBinary")
			}
			o.uArg = i
		case 5:
			var l int64
			if err := binary.Read(buf, binary.LittleEndian, &l); err != nil {
				return errors.Wrap(err, "failed to read length argument during UnmarshalBinary")
			}

			b := make([]byte, l)
			for i := int64(0); i < l; i++ {
				if err := binary.Read(buf, binary.LittleEndian, &b[i]); err != nil {
					return errors.Wrap(err, "failed to read bytes from buffer during UnmarshalBinary")
				}
			}
			o.uArg = b
		default:
			panic(fmt.Sprintf("Unknown tArg: %d", tArg))
		}
	}

	var hasComment int8
	if err := binary.Read(buf, binary.LittleEndian, &hasComment); err != nil {
		return errors.Wrap(err, "hasComment check failed during UnmarshalBinary")
	}

	if hasComment == 1 {
		if err := binary.Read(buf, binary.LittleEndian, &o.comment); err != nil {
			return errors.Wrap(err, "failed to read comment bytes during UnmarshalBinary")
		}
	}

	return nil
}

// Call executes the Op code in the context of given vm.State
func (o *op) Call(st *State) {
	o.OpHandler(st)
}

// SetArg sets the argument to this Op
func (o *op) SetArg(v interface{}) {
	o.uArg = v
}

func (o *op) SetComment(s string) {
	o.comment = s
}

// Arg returns the Op code's argument
func (o op) Arg() interface{} {
	return o.uArg
}

// ArgInt returns the integer representation of the argument
func (o op) ArgInt() int {
	v := interfaceToNumeric(o.uArg)
	return int(v.Int())
}

// ArgString returns the string representatin of the argument
func (o op) ArgString() string {
	// In most cases we do this because it's a sring
	if v, ok := o.uArg.(string); ok {
		return v
	}
	return interfaceToString(o.uArg)
}

func (o *op) String() string {
	buf := bytes.Buffer{}

	fmt.Fprintf(&buf, "Op[%s]", o.Type())

	if o.Type() == TXOPLoadLvar {
		n := o.uArg.(*node.LocalVarNode)
		fmt.Fprintf(&buf, " '%s' (%d)", n.Name, n.Offset)
	} else {
		if o.uArg != nil {
			fmt.Fprintf(&buf, " (%q)", o.ArgString())
		}
	}

	if v := o.comment; v != "" {
		fmt.Fprintf(&buf, " // %s", v)
	}

	return buf.String()
}
