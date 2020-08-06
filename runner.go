package coza

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type RequestParams struct {
	URL     string
	Timeout time.Duration
}

type Result interface {
	Duration() time.Duration
	NRead() int64
	Err() error
}

func RunWorkerPool(ctx context.Context, size int, in <-chan *RequestParams) <-chan Result {
	out := make(chan Result, size)

	wg := sync.WaitGroup{}

	for i := 0; i != size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			results := RunWorker(ctx, in)
			for r := range results {
				out <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// alternative implementation of RunWorkerPool.
// todo: compare performance
func _(ctx context.Context, size int, in <-chan *RequestParams) <-chan Result {
	out := make(chan Result, size)

	removeCase := func(cases []reflect.SelectCase, index int) []reflect.SelectCase {
		copy(cases[index:], cases[index+1:])
		return cases[:len(cases)-1]
	}

	go func() {
		cases := make([]reflect.SelectCase, size+1)

		for i := 0; i != size; i++ {
			resultC := RunWorker(ctx, in)
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(resultC),
			}
		}

		ctxCancelCase := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}

		cases[size] = ctxCancelCase

		for len(cases) != 1 {
			i, v, ok := reflect.Select(cases)
			if !ok {
				cases = removeCase(cases, i)

				continue
			}

			if &cases[i] == &ctxCancelCase {
				close(out)

				return
			}

			out <- v.Interface().(Result)
		}

		close(out)
	}()

	return out
}

func RunWorker(ctx context.Context, in <-chan *RequestParams) <-chan Result {
	out := make(chan Result)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(out)
				return

			case params, ok := <-in:
				if !ok {
					close(out)
					return
				}

				d, err := Run(ctx, http.DefaultClient, params)
				if err != nil {
					// todo: notify receiver
					log.Println(err)
					continue
				}

				out <- d
			}
		}
	}()

	return out
}

func Run(ctx context.Context, client *http.Client, params *RequestParams) (Result, error) {
	reqCtx, cancel := context.WithTimeout(ctx, params.Timeout)
	defer cancel()

	req, err := createRequest(reqCtx, params)
	if err != nil {
		return nil, fmt.Errorf("coza: run: %w", err)
	}

	duration, nRead, err := doRequest(client, req)
	if err != nil {
		err = fmt.Errorf("coza: run: %w", err)
	}

	result := resultImpl{
		duration: duration,
		nRead:    nRead,
		err:      err,
	}

	return &result, nil
}

func createRequest(ctx context.Context, params *RequestParams) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return req, nil
}

func doRequest(c *http.Client, req *http.Request) (time.Duration, int64, error) {
	start := time.Now()
	res, err := c.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, 0, err
	}

	defer func() {
		// todo: notify caller
		err = res.Body.Close()
		if err != nil {
			log.Printf("coza: request: unable to close response: %v\n", err.Error())
		}
	}()

	nRead, err := sizeOfBody(res.Body)
	if err != nil {
		return duration, nRead, fmt.Errorf("unable to do request: %w", err)
	}

	return duration, nRead, err
}

type nRead struct{}

func (nRead) Write(p []byte) (int, error) {
	return len(p), nil
}

func sizeOfBody(body io.Reader) (int64, error) {
	return io.Copy(nRead{}, body)
}

type resultImpl struct {
	duration time.Duration
	nRead    int64
	err      error
}

func (r *resultImpl) Duration() time.Duration {
	return r.duration
}

func (r *resultImpl) NRead() int64 {
	return r.nRead
}

func (r *resultImpl) Err() error {
	return r.err
}
