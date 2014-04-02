package main

import (
  "fmt"
  "os"
  "log"
  "time"
  "github.com/lestrrat/go-xslate"
  "strconv"
)

func main() {
  var err error

  var iter int64 = 100
  if len(os.Args) >= 2 {
    iter, err = strconv.ParseInt(os.Args[1], 10, 64)
    if err != nil {
      log.Fatalf("Failed to parse iter: %s", err)
    }
  }

  cache := true
  if len(os.Args) >= 3 {
    cache, err = strconv.ParseBool(os.Args[2])
    if err != nil {
      log.Fatalf("Failed to parse second argument: %s", err)
    }
  }

  var cacheLevel int
  if cache {
    cacheLevel = 1
  } else {
    cacheLevel = 0
  }

  tx, _ := xslate.New(xslate.Args {
    "Loader": xslate.Args {
      "CacheLevel": cacheLevel,
    },
  })

  file, err := os.OpenFile("/dev/null", os.O_WRONLY, 0777)
  if err != nil {
    log.Fatalf("Failed to open /dev/null: %s", err)
  }

  t0 := time.Now()
  for i := 0; i < int(iter); i++ {
    tx.RenderInto(file, "hello.tx", nil)
  }
  elapsed := time.Since(t0)

  fmt.Printf("* Elapsed %f secs\n", float64(elapsed.Nanoseconds()) / float64(time.Second))
  fmt.Printf("* Iter per sec: %f iter/sec\n", float64(iter) / float64(elapsed.Seconds()))
}