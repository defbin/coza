package coza

import (
	"math"
	"time"
)

type Stat interface {
	Duration() time.Duration
	Avg() time.Duration
	Min() time.Duration
	Median() time.Duration
	Max() time.Duration
	P90() time.Duration
	P95() time.Duration
	NRead() int64
}

type Metric interface {
	Duration() time.Duration
	NRead() int64
}

func Calc(metrics []Metric) Stat {
	stat := &statImpl{min: math.MaxInt64}

	if len(metrics) == 0 {
		return stat
	}

	p90i, p95i := percentIndex(len(metrics), 90), percentIndex(len(metrics), 95)

	for i, m := range metrics {
		d := m.Duration()
		stat.duration += d
		stat.nRead += m.NRead()

		if stat.min > d {
			stat.min = d
		}

		if stat.max < d {
			stat.max = d
		}

		if i == p90i {
			stat.p90 = stat.duration / time.Duration(i+1)
		}

		if i == p95i {
			stat.p95 = stat.duration / time.Duration(i+1)
		}
	}

	medianIndex := percentIndex(len(metrics), 50)
	stat.med = metrics[medianIndex].Duration()
	stat.avg = stat.duration / time.Duration(len(metrics))

	return stat
}

func percentIndex(l int, p float64) int {
	return int(math.Max(math.Round(float64(l)/100*p)-1, 0))
}

type statImpl struct {
	duration           time.Duration
	avg, min, med, max time.Duration
	p90, p95           time.Duration
	nRead              int64
}

func (s *statImpl) Duration() time.Duration {
	return s.duration
}

func (s *statImpl) Avg() time.Duration {
	return s.avg
}

func (s *statImpl) Min() time.Duration {
	return s.min
}

func (s *statImpl) Median() time.Duration {
	return s.med
}

func (s *statImpl) Max() time.Duration {
	return s.max
}

func (s *statImpl) P90() time.Duration {
	return s.p90
}

func (s *statImpl) P95() time.Duration {
	return s.p95
}

func (s *statImpl) NRead() int64 {
	return s.nRead
}
