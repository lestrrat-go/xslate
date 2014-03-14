package main

import (
  "flag"
  "fmt"
  "os"
  "github.com/lestrrat/go-xslate"
  "github.com/lestrrat/go-xslate/loader"
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
  // TODO: Accept --path arguments
  pwd, err := os.Getwd()
  if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to get current working directory: %s\n", err)
    os.Exit(1)
  }
  tx.Loader, _ = loader.NewLoadFile([]string { pwd })
  for _, file := range args {
    output, err := tx.Render(file, nil)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to render %s: %s\n", file, err)
      os.Exit(1)
    }
    fmt.Fprintf(os.Stdout, output)
  }
}