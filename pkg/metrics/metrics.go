package metrics

import (
	"time"
)

// Metric represents a statsite metric
type Metric interface {
	Emit()
}

// Timer Metric
// t := Timer(key)
// defer t.Emit()
type timer struct {
	key   string
	tags  []string
	start time.Time
	end   time.Time
}

func Timer(key string, tags ...string) *timer {
	return processor.NewTimer(key, time.Now(), tags...)
}

func (t *timer) Emit() {
	t.end = time.Now()
	processor.Publish(t)
}

// Counter Metric
// t := Counter(key)
// defer t.Emit()
type counter struct {
	key   string
	count int64
	tags  []string
}

func Counter(key string, tags ...string) *counter {
	return processor.NewCounter(key, 0, tags...)
}

func CounterAt(key string, i int, tags ...string) *counter {
	return processor.NewCounter(key, int64(i), tags...)
}

func CounterAt64(key string, i int64, tags ...string) *counter {
	return processor.NewCounter(key, i, tags...)
}

func (t *counter) Incr() {
	t.count += 1
}

func (t *counter) IncrBy(i int) {
	t.count += int64(i)
}

func (t *counter) IncrBy64(i int64) {
	t.count += i
}

func (t *counter) Emit() {
	processor.Publish(t)
}

type gauge struct {
	key   string
	value float64
	tags  []string
}

func Gauge(key string, tags ...string) *gauge {
	return processor.NewGauge(key, 0, tags...)
}

func GaugeAt(key string, value float64, tags ...string) *gauge {
	return processor.NewGauge(key, value, tags...)
}

func (t *gauge) Set(f float64) {
	t.value = f
}

func (t *gauge) Emit() {
	processor.Publish(t)
}
