package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type JSONProvider struct {
	params map[string]interface{}
}

func NewJSONProviderFromFile(filename string) (*JSONProvider, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewJSONProviderFromReader(f)
}

func NewJSONProviderFromString(data string) (*JSONProvider, error) {
	return NewJSONProviderFromReader(strings.NewReader(data))
}

func NewJSONProviderFromReader(r io.Reader) (*JSONProvider, error) {
	var params map[string]interface{}

	dec := json.NewDecoder(r)
	dec.UseNumber()
	if err := dec.Decode(&params); err != nil {
		return nil, err
	}

	return &JSONProvider{
		params: params,
	}, nil
}

func (t JSONProvider) Name() string {
	return "JSONProvider"
}

func (t JSONProvider) Update() error {
	return nil
}

func (t JSONProvider) GetInt(key string) (int, error) {
	n, err := t.get(key)
	if err != nil {
		return 0, err
	}

	i, err := n.Int64()
	if err != nil {
		return 0, NewInvalidParamError(t.Name(), key, err)
	}

	return int(i), nil
}

func (t JSONProvider) GetInt64(key string) (int64, error) {
	n, err := t.get(key)
	if err != nil {
		return 0, err
	}

	i, err := n.Int64()
	if err != nil {
		return 0, NewInvalidParamError(t.Name(), key, err)
	}

	return i, nil
}

func (t JSONProvider) GetFloat64(key string) (float64, error) {
	n, err := t.get(key)
	if err != nil {
		return 0, err
	}

	f, err := n.Float64()
	if err != nil {
		return 0, NewInvalidParamError(t.Name(), key, err)
	}

	return f, nil
}

func (t JSONProvider) GetString(key string) (string, error) {
	value, ok := t.params[key]
	if !ok {
		return "", NewNotFoundError(t.Name(), key)
	}

	s, ok := value.(string)
	if !ok {
		return "", NewInvalidParamError(t.Name(), key, fmt.Errorf("Not string"))
	}

	return s, nil
}

func (t JSONProvider) GetBool(key string) (bool, error) {
	value, ok := t.params[key]
	if !ok {
		return false, NewNotFoundError(t.Name(), key)
	}

	b, ok := value.(bool)
	if !ok {
		return false, NewInvalidParamError(t.Name(), key, fmt.Errorf("Not bool"))
	}

	return b, nil
}

func (t JSONProvider) get(key string) (json.Number, error) {
	value, ok := t.params[key]
	if !ok {
		return "", NewNotFoundError(t.Name(), key)
	}

	n, ok := value.(json.Number)
	if !ok {
		return "", NewInvalidParamErrorNAN(t.Name(), key, n)
	}

	return n, nil
}

type YAMLProvider struct {
	params map[string]interface{}
}

func NewYAMLProviderFromFile(filename string) (*YAMLProvider, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewYAMLProviderFromReader(f)
}

func NewYAMLProviderFromString(data string) (*YAMLProvider, error) {
	return NewYAMLProviderFromReader(strings.NewReader(data))
}

func NewYAMLProviderFromReader(r io.Reader) (*YAMLProvider, error) {
	var params map[string]interface{}
	if err := yaml.NewDecoder(r).Decode(&params); err != nil {
		return nil, err
	}

	return &YAMLProvider{
		params: params,
	}, nil
}

func (t YAMLProvider) Name() string {
	return "YAMLProvider"
}

func (t YAMLProvider) Update() error {
	return nil
}

func (t YAMLProvider) GetInt(key string) (int, error) {
	value, ok := t.params[key]
	if !ok {
		return 0, NewNotFoundError(t.Name(), key)
	}

	var i int
	switch v := value.(type) {
	case int:
		i = v
	case int64:
		i = int(v)
	default:
		return 0, NewInvalidParamErrorNAN(t.Name(), key, i)
	}

	return i, nil
}

func (t YAMLProvider) GetInt64(key string) (int64, error) {
	value, ok := t.params[key]
	if !ok {
		return 0, NewNotFoundError(t.Name(), key)
	}

	var i int64
	switch v := value.(type) {
	case int:
		i = int64(v)
	case int64:
		i = v
	default:
		return 0, NewInvalidParamErrorNAN(t.Name(), key, v)
	}

	return i, nil
}

func (t YAMLProvider) GetFloat64(key string) (float64, error) {
	value, ok := t.params[key]
	if !ok {
		return 0, NewNotFoundError(t.Name(), key)
	}

	var f float64
	switch v := value.(type) {
	case int:
		f = float64(v)
	case int64:
		f = float64(v)
	case float32:
		f = float64(v)
	case float64:
		f = v
	default:
		return 0, NewInvalidParamErrorNAN(t.Name(), key, v)
	}

	return f, nil
}

func (t YAMLProvider) GetString(key string) (string, error) {
	value, ok := t.params[key]
	if !ok {
		return "", NewNotFoundError(t.Name(), key)
	}

	s, ok := value.(string)
	if !ok {
		return "", NewInvalidParamErrorNAN(t.Name(), key, s)
	}

	return s, nil
}

func (t YAMLProvider) GetBool(key string) (bool, error) {
	value, ok := t.params[key]
	if !ok {
		return false, NewNotFoundError(t.Name(), key)
	}

	b, ok := value.(bool)
	if !ok {
		return false, NewInvalidParamErrorNAN(t.Name(), key, b)
	}

	return b, nil
}
