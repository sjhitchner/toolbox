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

func Initialize(done <-chan struct{}, backend Backend) {
	processor = NewProcessor(done, backend)
	processor.Loop()
}

type Backend interface {
	Timer(key string, dur time.Duration, tags ...string)
	Counter(key string, count int64, tags ...string)
	Gauge(key string, value float64, tags ...string)
}

type NopBackend struct {
}

func (t *NopBackend) Timer(key string, dur time.Duration, tags ...string) {
}

func (t *NopBackend) Counter(key string, count int64, tags ...string) {
}

func (t *NopBackend) Gauge(key string, value float64, tags ...string) {
}

type Processor struct {
	wg          sync.WaitGroup
	counterPool sync.Pool
	timerPool   sync.Pool
	gaugePool   sync.Pool

	backend Backend

	doneCh  <-chan struct{}
	queueCh chan Metric
}

func NewProcessor(done <-chan struct{}, backend Backend) *Processor {
	return &Processor{
		doneCh:  done,
		queueCh: make(chan Metric, BufferSize),
		backend: backend,
		counterPool: sync.Pool{
			New: func() interface{} {
				return &counter{}
			},
		},
		timerPool: sync.Pool{
			New: func() interface{} {
				return &timer{}
			},
		},
		gaugePool: sync.Pool{
			New: func() interface{} {
				return &gauge{}
			},
		},
	}
}

func (t *Processor) NewCounter(key string, count int64, tags ...string) *counter {
	m := t.counterPool.Get().(*counter)
	m.key = key
	m.count = count
	m.tags = tags
	return m
}

func (t *Processor) NewTimer(key string, start time.Time, tags ...string) *timer {
	m := t.timerPool.Get().(*timer)
	m.key = key
	m.start = start
	m.tags = tags
	return m
}

func (t *Processor) NewGauge(key string, value float64, tags ...string) *gauge {
	m := t.gaugePool.Get().(*gauge)
	m.key = key
	m.value = value
	m.tags = tags
	return m
}

func (t *Processor) Publish(metric Metric) {
	if t.backend == nil {
		return
	}

	select {
	case t.queueCh <- metric:
	default:
		// Channel is full so we are dropping metric
		switch m := metric.(type) {
		case *counter:
			t.counterPool.Put(m)

		case *timer:
			t.timerPool.Put(m)

		case *gauge:
			t.gaugePool.Put(m)

		default:
			log.Println("full")
		}
	}
}

func (t *Processor) Loop() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

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

func (t *Processor) innerLoop(metric Metric) {
	switch m := metric.(type) {
	case *counter:
		t.backend.Counter(m.key, m.count, m.tags...)
		t.counterPool.Put(m)

	case *timer:
		t.backend.Timer(m.key, m.end.Sub(m.start), m.tags...)
		t.timerPool.Put(m)

	case *gauge:
		t.backend.Gauge(m.key, m.value, m.tags...)
		t.gaugePool.Put(m)

	default:
		log.Println("invalid metric type")
	}
}
