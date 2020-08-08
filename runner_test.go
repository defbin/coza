package coza

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun(t *testing.T) {
	assertResultError := func(t *testing.T, got, want Result) {
		t.Helper()

		if !errors.Is(got.Err(), want.Err()) {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	server := httptest.NewServer(nil)
	defer server.Close()

	t.Run("success case", func(t *testing.T) {
		params := RequestParams{URL: server.URL}

		got, err := Run(context.Background(), http.DefaultClient, &params)
		want := &resultImpl{}

		assertNoError(t, err)
		assertResultError(t, got, want)
	})

	t.Run("empty url", func(t *testing.T) {
		params := RequestParams{}

		_, err := Run(context.Background(), http.DefaultClient, &params)
		want := ErrInvalidURL

		assertError(t, err, want)
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		params := RequestParams{URL: server.URL}

		cancel()

		got, err := Run(ctx, http.DefaultClient, &params)
		want := &resultImpl{err: context.Canceled}

		assertNoError(t, err)
		assertResultError(t, got, want)
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, _ := context.WithTimeout(context.Background(), 0)
		params := RequestParams{URL: server.URL}

		got, err := Run(ctx, http.DefaultClient, &params)
		want := &resultImpl{err: context.DeadlineExceeded}

		assertNoError(t, err)
		assertResultError(t, got, want)
	})
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err.Error())
	}
}

func assertError(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
