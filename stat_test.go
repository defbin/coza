package coza

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestCalc(t *testing.T) {
	makeRandomMetrics := func(n int) []Metric {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		metrics := make([]Metric, n)

		for i := 0; i != n; i++ {
			metrics[i] = &resultImpl{
				duration: time.Duration(r.Uint32()),
				nRead:    int64(r.Uint32()),
			}
		}

		return metrics
	}

	sumInt64 := func(xs []int64) int64 {
		var rv int64
		for _, d := range xs {
			rv += d
		}
		return rv
	}

	avgInt64 := func(xs []int64) int64 {
		return sumInt64(xs) / int64(len(xs))
	}

	minInt64 := func(xs []int64) int64 {
		var min int64 = math.MaxInt64
		for _, d := range xs {
			if min > d {
				min = d
			}
		}
		return min
	}

	maxInt64 := func(xs []int64) int64 {
		var max int64
		for _, d := range xs {
			if max < d {
				max = d
			}
		}
		return max
	}

	assertStat := func(t *testing.T, got, want Stat) {
		t.Helper()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v want %#v", got, want)
		}
	}

	t.Run("empty metrics", func(t *testing.T) {
		metrics := make([]Metric, 0)
		got := Calc(metrics)
		want := newStatImpl()

		assertStat(t, got, want)
	})

	t.Run("single metric", func(t *testing.T) {
		metric := &resultImpl{
			duration: 1,
			nRead:    2,
		}
		got := Calc([]Metric{metric})
		want := newStatImpl()
		want.duration = metric.duration
		want.avg = metric.duration
		want.min = metric.duration
		want.med = metric.duration
		want.max = metric.duration
		want.p90 = metric.duration
		want.p95 = metric.duration
		want.nRead = metric.nRead

		assertStat(t, got, want)
	})

	t.Run("many metric", func(t *testing.T) {
		metrics := makeRandomMetrics(20)
		got := Calc(metrics)

		durations, nRead := make([]int64, 20), make([]int64, 20)
		for i, m := range metrics {
			durations[i] = int64(m.Duration())
			nRead[i] = m.NRead()
		}

		want := &statImpl{
			duration: time.Duration(sumInt64(durations)),
			avg:      time.Duration(avgInt64(durations)),
			min:      time.Duration(minInt64(durations)),
			max:      time.Duration(maxInt64(durations)),
			med:      metrics[9].Duration(),
			p90:      time.Duration(avgInt64(durations[:18])),
			p95:      time.Duration(avgInt64(durations[:19])),
			nRead:    sumInt64(nRead),
		}

		assertStat(t, got, want)
	})
}
