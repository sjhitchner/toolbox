package config

import (
	"fmt"
	"time"
)

type Config interface {
	GetInt(key string) (int, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
}

// Provider
// Provide access to the underlying values
//
// TODO
// - Document base sources (JSON, ConfigMap, ConfigFiles)
// - Key/Value store
type Provider interface {
	Config
	Name() string
	Update() error
}

type Source struct {
	Provider Provider
	Interval time.Duration
	Priority int
}

type Sources []Source

func (t Sources) Len() int           { return len(t) }
func (t Sources) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Sources) Less(i, j int) bool { return t[i].Priority < t[j].Priority }

type ConfigError struct {
	Provider string
	Key      string
}

type NotFoundError struct {
	ConfigError
}

func NewNotFoundError(provider, key string) *NotFoundError {
	return &NotFoundError{
		ConfigError: ConfigError{
			Provider: provider,
			Key:      key,
		},
	}
}

func (t NotFoundError) Error() string {
	return fmt.Sprintf("%s (%s) key not found", t.Key, t.Provider)
}

type InvalidParamError struct {
	ConfigError
	Err error
}

func NewInvalidParamErrorNAN(provider, key string, v interface{}) *InvalidParamError {
	return NewInvalidParamError(provider, key, fmt.Errorf("NAN %v", v))
}

func NewInvalidParamError(provider, key string, err error) *InvalidParamError {
	return &InvalidParamError{
		ConfigError: ConfigError{
			Provider: provider,
			Key:      key,
		},
		Err: err,
	}
}

func (t InvalidParamError) Error() string {
	return fmt.Sprintf("%s (%s) key invalid %s", t.Key, t.Provider, t.Err.Error())
}



Thai Elephant Dung
Stargazer Cubensis
PE7 Isolated
Jedi Mind Fuck
Jack Frost
Hillbilly
White Rabbit
Tosohatchee
