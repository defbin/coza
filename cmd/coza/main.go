package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/defbin/coza"
)

var (
	tries          = flag.Int("r", 10, "number of total request")
	concurrency    = flag.Int("c", runtime.NumCPU(), "max number of concurrent requests")
	timeout        = flag.Duration("t", 1*time.Minute, "timeout for all requests")
	requestTimeout = flag.Duration("rt", 10*time.Second, "timeout per request")
)

func main() {
	flag.Parse()

	if nArg := flag.NArg(); nArg != 1 {
		log.Printf("Wrong number of argument %v. Expected 1.\n", nArg)
		os.Exit(1)
	}

	startedAt := time.Now()
	results := do()
	report(results)

	fmt.Printf("Completed in %v\n", time.Since(startedAt))
}

func do() []coza.Result {
	params := make(chan *coza.RequestParams, *concurrency)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)

	defer cancel()

	resultC := coza.RunWorkerPool(ctx, *concurrency, params)

	go func() {
		reqParams := &coza.RequestParams{
			URL:     "http://google.com",
			Timeout: *requestTimeout,
		}

		for i := 0; i != *tries; i++ {
			params <- reqParams
		}

		close(params)
	}()

	results := make([]coza.Result, 0, *tries)
	for r := range resultC {
		results = append(results, r)
	}

	return results
}

func report(results []coza.Result) {
	metrics := make([]coza.Metric, len(results))

	for i := range results {
		metrics[i] = results[i]
	}

	stat := coza.Calc(metrics)
	fmt.Println(formatStat(stat))
}

func formatStat(stat coza.Stat) string {
	return fmt.Sprintf(`
	Total: %v
	Average: %v
	Minimum: %v
	Median: %v
	Maximum: %v
	90%%: %v
	95%%: %v
	Bytes read: %v
	`,
		stat.Duration(),
		stat.Avg(),
		stat.Min(),
		stat.Median(),
		stat.Max(),
		stat.P90(),
		stat.P95(),
		stat.NRead())
}
