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
}

type NopBackend struct {
}

func (t *NopBackend) Timer(key string, dur time.Duration, tags ...string) {
}

func (t *NopBackend) Counter(key string, count int64, tags ...string) {
}

type Processor struct {
	wg          sync.WaitGroup
	counterPool sync.Pool
	timerPool   sync.Pool

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
	}
}

func (t *Processor) NewCounter(key string, count int64) *counter {
	m := t.counterPool.Get().(*counter)
	m.key = key
	m.count = count
	return m
}

func (t *Processor) NewTimer(key string, start time.Time) *timer {
	m := t.timerPool.Get().(*timer)
	m.key = key
	m.start = start
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
		t.backend.Counter(m.key, m.count)
		t.counterPool.Put(m)

	case *timer:
		t.backend.Timer(m.key, m.end.Sub(m.start))
		t.timerPool.Put(m)

	default:
		log.Println("invalid metric type")
	}
}
