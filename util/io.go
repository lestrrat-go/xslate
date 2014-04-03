package util

import (
  "bufio"
  "bytes"
  "io"
  "reflect"
)

var BufferedWriterType = reflect.TypeOf(bufio.NewWriter(&bytes.Buffer {}))

func IsBuffered(w io.Writer) bool {
  return reflect.TypeOf(w).String() == BufferedWriterType.String()
}

