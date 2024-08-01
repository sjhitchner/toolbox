package sqs

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type ErrorType int

const (
	UnmarshalError ErrorType = iota
	MarshalError
)

type Error struct {
	Msg types.Message
	Err error
}

func (t Error) Error() string {
	return t.Err.Error()
}

type Message[T any] struct {
	Msg types.Message
	Obj T
}

type Serializer[T any] interface {
	Unmarshal(in <-chan types.Message) (<-chan Message[T], <-chan error)
	Marshal(in <-chan T) (<-chan string, <-chan error)
}

type JSONSerializer[T any] struct {
}

func (t JSONSerializer[T]) Unmarshal(in <-chan types.Message) (<-chan Message[T], <-chan error) {
	outCh := make(chan Message[T])
	errCh := make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		for msg := range in {
			var obj T
			if err := json.Unmarshal([]byte(*msg.Body), &obj); err != nil {
				errCh <- err
			}

			outCh <- Message[T]{
				Msg: msg,
				Obj: obj,
			}
		}
	}()
	return outCh, errCh
}

func (t JSONSerializer[T]) Marshal(in <-chan T) (<-chan string, <-chan error) {
	outCh := make(chan string)
	errCh := make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		for obj := range in {
			b, err := json.Marshal(obj)
			if err != nil {
				errCh <- err
			}
			outCh <- string(b)
		}
	}()
	return outCh, errCh
}
