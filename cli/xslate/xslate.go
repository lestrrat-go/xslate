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

  // TODO: Accept --path arguments
  tx, err := xslate.New()
  if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to create Xslate instance: %s", err)
    os.Exit(1)
  }

  for _, file := range args {
    output, err := tx.Render(file, nil)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to render %s: %s\n", file, err)
      os.Exit(1)
    }
    fmt.Fprintf(os.Stdout, output)
  }
}