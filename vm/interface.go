package vm

import (
	"io"
	"reflect"
	"time"

	"github.com/lestrrat/go-xslate/internal/stack"
)

// ByteCode is the collection of op codes that the Xslate Virtual Machine
// should run. It is created from a compiler.Compiler
type ByteCode struct {
	OpList      []Op
	GeneratedOn time.Time
	Name        string
	Version     float32
}

// OpType is an integer identifying the type of op code
type OpType int

// OpHandler describes an op's actual code
type OpHandler func(*State)

// Op represents a single op. It has an OpType, OpHandler, and an optional
// parameter to be used
type Op interface {
	Arg() interface{}
	ArgString() string
	ArgInt() int
	Call(*State)
	Comment() string
	Handler() OpHandler
	SetArg(interface{})
	SetComment(string)
	String() string
	Type() OpType
}

type op struct {
	OpType
	OpHandler
	uArg    interface{}
	comment string
}

// State keeps track of Xslate Virtual Machine state
type State struct {
	opidx int
	pc    *ByteCode

	stack     stack.Stack
	markstack stack.Stack

	// output
	output io.Writer
	warn   io.Writer

	// template variables
	vars Vars

	// registers
	sa   interface{}
	sb   interface{}
	targ interface{}

	// Stack used by frames
	framestack stack.Stack
	frames     stack.Stack

	Loader       byteCodeLoader
	MaxLoopCount int
}

// LoopVar is the variable available within FOREACH loops
type LoopVar struct {
	Index    int           // 0 origin, current index
	Count    int           // loop.Index + 1
	Body     reflect.Value // alias to array
	Size     int           // len(loop.Body)
	MaxIndex int           // loop.Size - 1
	PeekNext interface{}   // previous item. nil if not available
	PeekPrev interface{}   // next item. nil if not available
	IsFirst  bool          // true only if Index == 0
	IsLast   bool          // true only if Index == MaxIndex
}

// Vars represents the variables passed into the Virtual Machine
type Vars map[string]interface{}

// This interface exists solely to avoid importing loader.ByteCodeLoader
// which is a cause for import loop
type byteCodeLoader interface {
	Load(string) (*ByteCode, error)
}

// VM represents the Xslate Virtual Machine
type VM struct {
	st        *State
	functions Vars
	Loader    byteCodeLoader
}

// These TXOP... constants are identifiers for each op
const (
	TXOPNoop OpType = iota
	TXOPNil
	TXOPMoveToSb
	TXOPMoveFromSb
	TXOPLiteral
	TXOPFetchSymbol
	TXOPFetchFieldSymbol
	TXOPFetchArrayElement
	TXOPMarkRaw
	TXOPUnmarkRaw
	TXOPPrint
	TXOPPrintRaw
	TXOPPrintRawConst
	TXOPSaveToLvar
	TXOPLoadLvar
	TXOPAdd
	TXOPSub
	TXOPMul
	TXOPDiv
	TXOPAnd
	TXOPGoto
	TXOPForStart
	TXOPForIter
	TXOPHTMLEscape
	TXOPUriEscape
	TXOPEquals
	TXOPNotEquals
	TXOPLessThan
	TXOPGreaterThan
	TXOPPopmark
	TXOPPushmark
	TXOPPopFrame
	TXOPPushFrame
	TXOPPush
	TXOPPop
	TXOPFunCall
	TXOPFunCallSymbol
	TXOPFunCallOmni
	TXOPMethodCall
	TXOPRange
	TXOPMakeArray
	TXOPMakeHash
	TXOPInclude
	TXOPWrapper
	TXOPFilter
	TXOPSaveWriter
	TXOPRestoreWriter
	TXOPEnd
	TXOPMax
)

var opnames = make([]string, TXOPMax)
var ophandlers = make([]OpHandler, TXOPMax)
