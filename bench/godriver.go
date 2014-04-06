package main

import (
  "bufio"
  "io/ioutil"
  "flag"
  "fmt"
  "time"
  "github.com/lestrrat/go-xslate"
  prof "github.com/davecheney/profile"
)

var iter = flag.Int("iterations", 100, "Number of iterations to perform")
var cacheLevel = flag.Int("cache", 1, "Use of cache (0=no cache, 1=cache + verify freshness, 2=always use cache)")
var profile = flag.Bool("profile", false, "Generate profile")

func main() {
  flag.Parse()

  tx, _ := xslate.New(xslate.Args {
    "Loader": xslate.Args {
      "CacheLevel": *cacheLevel,
    },
  })

  t0 := time.Now()

  if *profile {
    config := prof.Config {
      CPUProfile: true,
      MemProfile: true,
      BlockProfile: true,
      ProfilePath: ".",
    }
    defer prof.Start(&config).Stop()
  }

  f := bufio.NewWriter(ioutil.Discard)
  for i := 0; i < *iter; i++ {
    tx.RenderInto(f, "hello.tx", nil)
  }

  elapsed := time.Since(t0)

  fmt.Printf("* Elapsed %f secs\n", float64(elapsed.Seconds()))
  fmt.Printf("* Secs per iter: %f sec/iter\n", float64(elapsed.Seconds()) / float64(*iter))
  fmt.Printf("* Iter per sec: %f iter/sec\n", float64(*iter) / float64(elapsed.Seconds()))
}