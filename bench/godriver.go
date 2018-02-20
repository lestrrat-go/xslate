package main

import (
	"bufio"
	"flag"
	"fmt"
	prof "github.com/davecheney/profile"
	"github.com/lestrrat-go/xslate"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var iter = flag.Int("iterations", 100, "Number of iterations to perform")
var cacheLevel = flag.Int("cache", 1, "Use of cache (0=no cache, 1=cache + verify freshness, 2=always use cache)")
var profile = flag.Bool("profile", false, "Generate profile")
var template = flag.String("template", "hello.tx", "Template file to use")
var stdout = flag.Bool("stdout", false, "Direct output to stdout")

func main() {
	flag.Parse()

	tx, _ := xslate.New(xslate.Args{
		"Loader": xslate.Args{
			"CacheLevel": *cacheLevel,
		},
	})

	t0 := time.Now()

	if *profile {
		config := prof.Config{
			CPUProfile:   true,
			MemProfile:   true,
			BlockProfile: true,
			ProfilePath:  ".",
		}
		defer prof.Start(&config).Stop()
	}

	var output io.Writer
	if *stdout {
		output = os.Stdout
	} else {
		output = ioutil.Discard
	}
	f := bufio.NewWriter(output)

	for i := 0; i < *iter; i++ {
		err := tx.RenderInto(f, *template, nil)
		if err != nil {
			io.WriteString(os.Stderr, err.Error())
		}
		f.Flush()
	}

	elapsed := time.Since(t0)

	fmt.Printf("* Elapsed %f secs\n", float64(elapsed.Seconds()))
	fmt.Printf("* Secs per iter: %f sec/iter\n", float64(elapsed.Seconds())/float64(*iter))
	fmt.Printf("* Iter per sec: %f iter/sec\n", float64(*iter)/float64(elapsed.Seconds()))
}
