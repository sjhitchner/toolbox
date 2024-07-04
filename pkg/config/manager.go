package config

import (
	"sort"
	"time"
)

type Manager struct {
	sources Sources
	done    chan struct{}
	context Config
}

func New() *Manager {
	return &Manager{}
}

func (t *Manager) AddProvider(provider Provider, priority int, interval time.Duration) {
	t.sources = append(t.sources, Source{
		Provider: provider,
		Interval: interval,
		Priority: priority,
	})
}

func (t *Manager) Start() error {
	sort.Sort(t.sources)

	t.done = make(chan struct{})

	for _, source := range t.sources {
		t.startSourceUpdate(source)
	}

	return nil
}

func (t *Manager) startSourceUpdate(source Source) {
	// TODO Error handling
	source.Provider.Update()

	for {
		select {
		case <-t.done:
			return
		case <-time.After(source.Interval):
		}

		source.Provider.Update()
	}
}

func (t Manager) GetInt(key string, dflt int) int {
	val, err := t.context.GetInt(key)
	if err != nil {
		// TOOD log
		return dflt
	}
	return val
}

func (t Manager) GetInt64(key string, dflt int64) int64 {
	val, err := t.context.GetInt64(key)
	if err != nil {
		// TOOD log
		return dflt
	}
	return val
}

func (t Manager) GetFloat64(key string, dflt float64) float64 {
	val, err := t.context.GetFloat64(key)
	if err != nil {
		// TOOD log
		return dflt
	}
	return val
}

func (t Manager) GetString(key, dflt string) string {
	val, err := t.context.GetString(key)
	if err != nil {
		// TOOD log
		return dflt
	}
	return val
}

func (t Manager) GetBool(key string, dflt bool) bool {
	val, err := t.context.GetBool(key)
	if err != nil {
		// TOOD log
		return dflt
	}
	return val
}
