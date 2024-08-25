package sqs

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sjhitchner/toolbox/pkg/metrics"
	"github.com/sjhitchner/toolbox/pkg/streaming"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go/aws"
)

const (
	BackoffFactor = 1.5
)

func (t *SQSClient[T]) Send(workerCount int, inCh <-chan T) <-chan error {
	serialChs := make([]<-chan string, workerCount)
	errChs := make([]<-chan error, 2*workerCount)

	for i := 0; i < workerCount; i++ {
		serialChs[i], errChs[i] = t.serializer.Marshal(inCh)
	}

	sendCh := streaming.MergeBuffer(t.BufferSize, serialChs...)

	for i := 0; i < workerCount; i++ {
		errChs[i+workerCount] = t.sendLoop(sendCh)
	}

	return streaming.MergeDone(t.doneCh, errChs...)
}

func (t *SQSClient[T]) sendLoop(inCh <-chan string) <-chan error {
	errCh := make(chan error)

	retryCh := make(chan string)

	go func() {
		defer close(errCh)

		for {
			select {
			case <-t.doneCh:
				return
			case message, ok := <-inCh:
				if !ok {
					return
				}

				// TODO timeout
				ctx := context.Background()
				if err := t.send(ctx, message); err != nil {
					errCh <- err
				}
			}
		}
	}()
	return errCh
}

func (t *SQSClient[T]) send(ctx context.Context, message string) error {
	counter := metrics.Counter("sqs_send_count")
	errCounter := metrics.Counter("sqs_send_error")
	defer counter.Emit()
	defer errCounter.Emit()
	defer metrics.Timer("sqs_send_duration").Emit()

	request := sqs.SendMessageInput{
		QueueUrl:    &t.queueURL,
		MessageBody: &message,
	}

	_, err := t.client.SendMessage(ctx, &request)
	if err != nil {
		errCounter.Incr()
		return &Error{
			Type: SendError,
			Err:  err,
		}
	}

	counter.Incr()
	return nil
}

func (t *SQSClient[T]) SendBatch(workerCount int, inCh <-chan T) <-chan error {
	serialChs := make([]<-chan string, workerCount)
	errChs := make([]<-chan error, 2*workerCount)

	for i := 0; i < workerCount; i++ {
		serialChs[i], errChs[i] = t.serializer.Marshal(inCh)
	}

	sendCh := streaming.MergeBuffer(t.BufferSize, serialChs...)

	for i := 0; i < workerCount; i++ {
		errChs[i+workerCount] = t.batchSendLoop(sendCh)
	}

	return streaming.MergeDone(t.doneCh, errChs...)
}

func (t *SQSClient[T]) batchSendLoop(inCh <-chan string) <-chan error {
	errCh := make(chan error)

	retryCh := make(chan string)

	go func() {
		defer close(errCh)
		defer close(retryCh)

		var errCount float64

		for {
			select {
			case <-t.doneCh:
				return

			case messages, ok := <-streaming.Batch(t.doneCh, inCh, t.BatchSize, t.BatchTimeout):
				if !ok {
					return
				}

				if err := t.batchSend(messages); err != nil {
					for _, message := range messages {
						retryCh <- message
					}
					errCh <- err
				}

			case messages := <-streaming.Batch(t.doneCh, retryCh, t.BatchSize, t.BatchTimeout):
				if err := t.batchSend(messages); err != nil {
					for _, message := range messages {
						retryCh <- message
					}
					errCh <- err

					errCount++

					backoff := 100 * math.Pow(BackoffFactor, errCount)
					<-time.After(time.Duration(backoff) * time.Millisecond)
				} else {
					errCount = 0
				}
			}
		}
	}()
	return errCh
}

func (t *SQSClient[T]) batchSend(messages []string) error {
	if len(messages) == 0 {
		return nil
	}

	entries := make([]types.SendMessageBatchRequestEntry, 0, 10)
	for i := 0; i < len(messages); i += 10 {
		entries = append(entries, types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("msg_%d", i)),
			MessageBody: aws.String(fmt.Sprintf("This is message %d", i)),
		})
	}

	// TODO timeout
	ctx := context.Background()
	return t.batchSQSSend(ctx, entries)
}

func (t *SQSClient[T]) batchSQSSend(ctx context.Context, entries []types.SendMessageBatchRequestEntry) error {
	counter := metrics.Counter("sqs_send_batch_count")
	errCounter := metrics.Counter("sqs_send_batch_error")
	defer counter.Emit()
	defer errCounter.Emit()
	defer metrics.Timer("sqs_send_batch_duration").Emit()

	request := sqs.SendMessageBatchInput{
		QueueUrl: &t.queueURL,
		Entries:  entries,
	}

	_, err := t.client.SendMessageBatch(ctx, &request)
	if err != nil {
		errCounter.Incr()
		return &Error{
			Type: SendError,
			Err:  err,
		}
	}

	counter.Incr()
	return nil
}
