// Zero-allocatio metric library
// Handy functions for adding metrics easier
package metrics

import (
	"log"
	"sync"
	"time"
)

const (
	DefaultBufferSize = 8096
)

var (
	BufferSize = DefaultBufferSize
	processor  = NewProcessor(nil, nil)
)

// Initialize
// Call  to initialize
func Initialize(done <-chan struct{}, backend Backend) {
	processor = NewProcessor(done, backend)
	processor.Loop()
}

// Backend
// Implement a backend to support a custom metric backend
type Backend interface {
	Timer(key string, dur time.Duration, tags ...string)
	Counter(key string, count int64, tags ...string)
	Gauge(key string, value float64, tags ...string)
	Histogram(key string, value float64, tags ...string)
	Distribution(key string, value float64, tags ...string)
}

// NopBackend
// Metrics are ignored
type NopBackend struct {
}

func (t *NopBackend) Timer(key string, dur time.Duration, tags ...string) {
}

func (t *NopBackend) Counter(key string, count int64, tags ...string) {
}

func (t *NopBackend) Gauge(key string, value float64, tags ...string) {
}

// Processor
type Processor struct {
	pool sync.Pool

	backend Backend

	doneCh  <-chan struct{}
	queueCh chan *metric
}

func NewProcessor(done <-chan struct{}, backend Backend) *Processor {
	return &Processor{
		doneCh:  done,
		queueCh: make(chan *metric, BufferSize),
		backend: backend,
		pool: sync.Pool{
			New: func() interface{} {
				return &metric{}
			},
		},
	}
}

func (t *Processor) NewCounter(key string, count int64, tags ...string) CounterMetric {
	m := t.pool.Get().(*metric)
	m.typ = CounterType
	m.key = key
	m.count = count
	m.tags = tags
	return m
}

func (t *Processor) NewTimer(key string, start time.Time, tags ...string) TimerMetric {
	m := t.pool.Get().(*metric)
	m.typ = TimerType
	m.key = key
	m.start = start
	m.tags = tags
	return m
}

func (t *Processor) NewGauge(key string, value float64, tags ...string) GaugeMetric {
	m := t.pool.Get().(*metric)
	m.typ = GaugeType
	m.key = key
	m.value = value
	m.tags = tags
	return m
}

func (t *Processor) NewHistogram(key string, value float64, tags ...string) HistogramMetric {
	m := t.pool.Get().(*metric)
	m.typ = HistogramType
	m.key = key
	m.value = value
	m.tags = tags
	return m
}

func (t *Processor) NewDistribution(key string, value float64, tags ...string) DistributionMetric {
	m := t.pool.Get().(*metric)
	m.typ = DistributionType
	m.key = key
	m.value = value
	m.tags = tags
	return m
}

func (t *Processor) Publish(metric *metric) {
	if t.backend == nil {
		return
	}

	select {
	case t.queueCh <- metric:
	default:
		t.pool.Put(metric)
		// TODO logging
		log.Println("metric queue full")
	}
}

func (t *Processor) Loop() {
	go func() {
		for {
			select {
			case <-t.doneCh:
				return
			case metric := <-t.queueCh:
				t.innerLoop(metric)
			}
		}
	}()
}

func (t *Processor) innerLoop(m *metric) {
	switch m.typ {
	case CounterType:
		t.backend.Counter(m.key, m.count, m.tags...)

	case TimerType:
		t.backend.Timer(m.key, m.end.Sub(m.start), m.tags...)

	case GaugeType:
		t.backend.Gauge(m.key, m.value, m.tags...)

	case HistogramType:
		t.backend.Histogram(m.key, m.value, m.tags...)

	case DistributionType:
		t.backend.Distribution(m.key, m.value, m.tags...)

	default:
		log.Println("invalid metric type")
	}

	t.pool.Put(m)
}
