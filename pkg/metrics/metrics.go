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
	start time.Time
	end   time.Time
}

func Timer(key string) *timer {
	// return &timer{key: key, start: time.Now()}
	return processor.NewTimer(key, time.Now())
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
}

func Counter(key string) *counter {
	return processor.NewCounter(key, 0)
}

func CounterAt(key string, i int) *counter {
	// return &counter{key, int64(i)}
	return processor.NewCounter(key, int64(i))
}

func CounterAt64(key string, i int64) *counter {
	return processor.NewCounter(key, i)
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

/*
type timerCounter struct {
	timer   *timer
	counter *counter
}

func TimerCounter(key string) *timerCounter {
	return &timerCounter{
		Timer(key),
		CounterAt(key, 1),
	}
}

func TimerCounterAt(key string, i int) *timerCounter {
	return &timerCounter{
		Timer(key),
		CounterAt(key, i),
	}
}

func (t *timerCounter) Incr() {
	t.counter.Incr()
}

func (t *timerCounter) IncrBy(i int) {
	t.counter.IncrBy(i)
}

func (t *timerCounter) Emit() {
	t.counter.Emit()
	t.timer.Emit()
}

type keyvalue struct {
	key   string
	value string
}

func KeyValue(key string, value string) *keyvalue {
	return &keyvalue{key, value}
}

func (t *keyvalue) Emit() {
	kv := NewKeyValue(t.key, t.value)
	processor.Publish(kv)
}

type gauge struct {
	key   string
	value int
}

func Gauge(key string) *gauge {
	return &gauge{key, 0}
}

func GaugeAt(key string, value int) *gauge {
	return &gauge{key, value}
}

func (t *gauge) Incr() {
	t.value += 1
}

func (t *gauge) IncrBy(i int) {
	t.value += i
}

func (t *gauge) Emit() {
	gauge := NewGauge(t.key, t.value)
	processor.Publish(gauge)
}
*/
