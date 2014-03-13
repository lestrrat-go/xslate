package main

import (
  "flag"
  "fmt"
  "os"
  "github.com/lestrrat/go-xslate"
)

func usage() {
  fmt.Fprintf(os.Stderr, "usage: xslate [options...] [input-files]\n")
  flag.PrintDefaults()
  os.Exit(2)
}

func main() {
  flag.Usage = usage
  flag.Parse()

  args := flag.Args()
  if len(args) < 1 {
    fmt.Fprintf(os.Stderr, "Input file is missing.\n")
    os.Exit(1)
  }

  tx := xslate.New()
  for _, file := range args {
    fh, err := os.Open(file)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to open %s for reading: %s\n", file, err)
      os.Exit(1)
    }

    output, err := tx.RenderReader(fh, nil)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to render %s: %s\n", file, err)
      os.Exit(1)
    }
    fmt.Fprintf(os.Stdout, output)
  }
}