package metrics

import (
	"time"
)

type MetricType int

const (
	CounterType MetricType = iota
	GaugeType
	TimerType
	HistogramType
	DistributionType
)

// Metric represents a statsite metric
type Metric interface {
	Emit()
}

// Counter Metric
// t := Counter(key)
// defer t.Emit()
type CounterMetric interface {
	Metric
	Incr()
	IncrBy(int)
	IncrBy64(int64)
}

type GaugeMetric interface {
	Metric
	Set(float64)
}

type DistributionMetric interface {
	Metric
	Set(float64)
}

type HistogramMetric interface {
	Metric
	Set(float64)
}

// Timer Metric
// t := Timer(key)
// defer t.Emit()
type TimerMetric interface {
	Metric
}

type metric struct {
	typ   MetricType
	key   string
	tags  []string
	start time.Time
	end   time.Time
	count int64
	value float64
}

func (t *metric) Incr() {
	t.count += 1
}

func (t *metric) IncrBy(i int) {
	t.count += int64(i)
}

func (t *metric) IncrBy64(i int64) {
	t.count += i
}

func (t *metric) Set(f float64) {
	t.value = f
}

func (t *metric) Emit() {
	t.end = time.Now()
	processor.Publish(t)
}

func Timer(key string, tags ...string) TimerMetric {
	return processor.NewTimer(key, time.Now(), tags...)
}

func Counter(key string, tags ...string) CounterMetric {
	return processor.NewCounter(key, 0, tags...)
}

func CounterAt(key string, i int, tags ...string) CounterMetric {
	return processor.NewCounter(key, int64(i), tags...)
}

func CounterAt64(key string, i int64, tags ...string) CounterMetric {
	return processor.NewCounter(key, i, tags...)
}

func Gauge(key string, tags ...string) GaugeMetric {
	return processor.NewGauge(key, 0, tags...)
}

func GaugeAt(key string, value float64, tags ...string) GaugeMetric {
	return processor.NewGauge(key, value, tags...)
}

func Histogram(key string, tags ...string) HistogramMetric {
	return processor.NewHistogram(key, 0, tags...)
}

func HistogramAt(key string, value float64, tags ...string) HistogramMetric {
	return processor.NewHistogram(key, value, tags...)
}

func Distribution(key string, tags ...string) DistributionMetric {
	return processor.NewDistribution(key, 0, tags...)
}

func DistributionAt(key string, value float64, tags ...string) DistributionMetric {
	return processor.NewDistribution(key, value, tags...)
}
