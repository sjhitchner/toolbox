package datadog

import (
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

const (
	DefaultHostname = "127.0.0.1:8125"
)

type DatadogBackend struct {
	client *statsd.Client
}

func New(hostname, namespace string, tags ...string) (*DatadogBackend, error) {
	client, err := statsd.New(hostname,
		statsd.WithNamespace(namespace),
		statsd.WithTags(tags),
	)
	if err != nil {
		return nil, err
	}

	return &DatadogBackend{
		client: client,
	}, nil
}

func (t *DatadogBackend) Timer(key string, dur time.Duration, tags ...string) {
	if err := t.client.Timing(key, dur, formatTags(tags), 1); err != nil {
		// TODO Logging
		log.Println(err)
	}
}

func (t *DatadogBackend) Counter(key string, count int64, tags ...string) {
	if err := t.client.Count(key, count, formatTags(tags), 1); err != nil {
		log.Println(err)
	}
}

func (t *DatadogBackend) Gauge(key string, value float64, tags ...string) {
	if err := t.client.Gauge(key, value, formatTags(tags), 1); err != nil {
		log.Println(err)
	}
}

func (t *DatadogBackend) Histogram(key string, value float64, tags ...string) {
	if err := t.client.Histogram(key, value, formatTags(tags), 1); err != nil {
		log.Println(err)
	}
}

func (t *DatadogBackend) Distribution(key string, value float64, tags ...string) {
	if err := t.client.Distribution(key, value, formatTags(tags), 1); err != nil {
		log.Println(err)
	}
}

func formatTags(tag []string) []string {
	if len(tag) < 2 {
		return nil
	}

	arr := make([]string, 0, len(tag)/2)
	for i := 1; i < len(tag); i += 2 {
		arr = append(arr, tag[i-1]+":"+tag[i])
	}

	return arr
}
