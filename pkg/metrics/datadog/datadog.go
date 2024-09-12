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
	if err := t.client.Timing(key, dur, tags, 1); err != nil {
		log.Println(err)
	}
}

func (t *DatadogBackend) Counter(key string, count int64, tags ...string) {
	if err := t.client.Count(key, count, tags, 1); err != nil {
		log.Println(err)
	}
}
