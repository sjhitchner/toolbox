package sqs

import (
	"context"
	"fmt"

	"github.com/sjhitchner/toolbox/pkg/metrics"
	"github.com/sjhitchner/toolbox/pkg/streaming"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func (t *SQSClient[T]) StartSending(workerCount int, inCh <-chan T) <-chan error {
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
		return err
	}

	fmt.Println("sent")

	counter.Incr()
	return nil
}
