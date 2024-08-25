package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/sjhitchner/toolbox/pkg/metrics"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSClient[T any] struct {
	client       *sqs.Client
	queueURL     string
	BufferSize   int
	BatchSize    int
	BatchTimeout time.Duration
	doneCh       <-chan struct{}
	serializer   Serializer[T]
}

func New[T any](done <-chan struct{}, cfg aws.Config, serializer Serializer[T], queueURL string) (*SQSClient[T], error) {
	client := sqs.NewFromConfig(cfg)

	return &SQSClient[T]{
		client:     client,
		queueURL:   queueURL,
		doneCh:     done,
		BatchSize:  10,
		BufferSize: 100,
		serializer: serializer,
	}, nil
}

func NewJSONQueue[T any](done <-chan struct{}, cfg aws.Config, queueURL string) (*SQSClient[T], error) {
	client := sqs.NewFromConfig(cfg)

	return &SQSClient[T]{
		client:     client,
		queueURL:   queueURL,
		doneCh:     done,
		BatchSize:  10,
		BufferSize: 100,
		serializer: JSONSerializer[T]{},
	}, nil
}

func (t *SQSClient[T]) DeleteMessage(message types.Message) error {
	counter := metrics.Counter("sqs_delete_count")
	errCounter := metrics.Counter("sqs_delete_error")
	defer counter.Emit()
	defer errCounter.Emit()
	defer metrics.Timer("sqs_delete_duration").Emit()

	ctx := context.Background()
	_, err := t.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &t.queueURL,
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		errCounter.Incr()
		return fmt.Errorf("Error deleting message: %v", err)
	}

	counter.Incr()
	return nil
}
