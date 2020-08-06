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

type AppConfig struct {
	tries          int
	concurrency    int
	timeout        time.Duration
	requestTimeout time.Duration
	url            string
}

var appConfig AppConfig

func init() {
	flag.IntVar(&appConfig.tries, "r", 10, "number of total request")
	flag.IntVar(&appConfig.concurrency, "c", runtime.NumCPU(), "max number of concurrent requests")
	flag.DurationVar(&appConfig.timeout, "t", 1*time.Minute, "timeout for all requests")
	flag.DurationVar(&appConfig.requestTimeout, "rt", 10*time.Second, "timeout per request")
	flag.Parse()

	if nArg := flag.NArg(); nArg != 1 {
		log.Printf("Wrong number of argument: %v. Expected 1.\n", nArg)
		os.Exit(1)
	}

	appConfig.url = flag.Arg(0)
}

func main() {
	startedAt := time.Now()

	results := do(appConfig)
	fmt.Println(report(results))

	fmt.Printf("Completed in %v\n", time.Since(startedAt))
}

func do(config AppConfig) []coza.Result {
	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()

	paramsC := make(chan *coza.RequestParams, config.concurrency)
	resultC := coza.RunWorkerPool(ctx, config.concurrency, paramsC)

	go func() {
		params := &coza.RequestParams{
			URL:     config.url,
			Timeout: config.requestTimeout,
		}

		for i := 0; i != config.tries; i++ {
			paramsC <- params
		}

		close(paramsC)
	}()

	results := make([]coza.Result, 0, config.tries)
	for r := range resultC {
		results = append(results, r)
	}

	return results
}

func report(results []coza.Result) string {
	metrics := make([]coza.Metric, len(results))

	for i := range results {
		metrics[i] = results[i]
	}

	stat := coza.Calc(metrics)
	return formatStat(stat)
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
