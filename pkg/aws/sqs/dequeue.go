package sqs

import (
	"context"
	"log"
	"time"

	"github.com/sjhitchner/toolbox/pkg/metrics"
	"github.com/sjhitchner/toolbox/pkg/streaming"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	MaxWaitTime = 127
)

func (t *SQSClient[T]) StartPolling(workerCount int) (<-chan Message[T], <-chan error) {
	pollChs := make([]<-chan types.Message, workerCount)
	serialChs := make([]<-chan Message[T], workerCount)
	errChs := make([]<-chan error, 2*workerCount)

	for i := 0; i < workerCount; i++ {
		pollChs[i], errChs[i] = t.receiveLoop()
	}

	msgCh := streaming.Merge(pollChs...)

	for i := 0; i < workerCount; i++ {
		serialChs[i], errChs[workerCount+i] = t.serializer.Unmarshal(msgCh)
	}

	outCh := streaming.Merge(serialChs...)
	errCh := streaming.Merge(errChs...)

	return outCh, errCh
}

func (t *SQSClient[T]) receiveLoop() (<-chan types.Message, <-chan error) {
	outCh := make(chan types.Message, t.BufferSize)
	errCh := make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		waitTime := time.Nanosecond

		for {
			select {
			case <-t.doneCh:
				return

			case <-time.After(waitTime):
				// TODO timeout
				ctx := context.Background()
				messages, err := t.receive(ctx)
				if err != nil {
					errCh <- err
				}

				for _, message := range messages {
					outCh <- message
				}

				if len(messages) > 0 {
					waitTime = time.Nanosecond
					continue
				}

				// Backoff and sleep to avoid excessive API calls
				waitTime = (waitTime * 2) + time.Second
				if waitTime > MaxWaitTime*time.Second {
					waitTime = MaxWaitTime * time.Second
				}
				log.Printf("No messages waiting %s.....\n", waitTime)
			}
		}
	}()
	return outCh, errCh
}

func (t *SQSClient[T]) receive(ctx context.Context) ([]types.Message, error) {
	counter := metrics.Counter("sqs_receive_count")
	errCounter := metrics.Counter("sqs_receive_error")
	defer counter.Emit()
	defer errCounter.Emit()
	defer metrics.Timer("sqs_receive_duration").Emit()

	request := sqs.ReceiveMessageInput{
		QueueUrl:            &t.queueURL,
		MaxNumberOfMessages: t.BatchSize,
		WaitTimeSeconds:     10,
	}

	resp, err := t.client.ReceiveMessage(ctx, &request)
	if err != nil {
		errCounter.Incr()
		return nil, &Error{
			Type: ReceiveError,
			Err:  err,
		}
	}

	counter.IncrBy(len(resp.Messages))

	return resp.Messages, nil
}

/*

receiverFn := func() {

		var noMessages int
		for {
			select {
			case <-doneCh:
				return
			default:

				ctx := context.Background()
				numReceived, err := t.receiveMessages(ctx, outCh)
				if err != nil {
					errCh <- err
				}

			}
		}
	}


func (t *SQSClient) receiveMessages(ctx context.Context, inCh chan<- types.Message) (int, error) {
	counter := metrics.Counter("sqs_receive_count")
	errCounter := metrics.Counter("sqs_receive_error")
	defer counter.Emit()
	defer errCounter.Emit()
	defer metrics.Timer("sqs_receive_duration").Emit()

	request := sqs.ReceiveMessageInput{
		QueueUrl:            &t.queueURL,
		MaxNumberOfMessages: t.BatchSize,
		WaitTimeSeconds:     10,
	}

	resp, err := t.client.ReceiveMessage(ctx, &request)
	if err != nil {
		errCounter.Incr()
		return 0, fmt.Errorf("Error receiving messages: %v", err)
	}

	counter.IncrBy(len(resp.Messages))

	for _, message := range resp.Messages {
		inCh <- message
	}

	return len(resp.Messages), nil
}


func main() {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable is required")
	}

	// Initialize StatsD
	sink, err := statsd.NewStatsdSink("localhost:8125")
	if err != nil {
		log.Fatalf("Error creating StatsD sink: %v", err)
	}
	metrics.NewGlobal(metrics.DefaultConfig("sqs_client"), sink)

	doneCh := make(chan struct{})

	sqsClient, err := NewSQSClient(queueURL, messageCh, doneCh)
	if err != nil {
		log.Fatalf("Error creating SQS client: %v", err)
	}

	workerCount := 5 // Number of concurrent workers
	go sqsClient.StartPolling(workerCount)

	// Example of processing messages from the channel
	go func() {
		for message := range messageCh {
			fmt.Printf("Received message: %s\n", *message.Body)
			// Process the message
			metrics.IncrCounter([]string{"messages_processed"}, 1)
		}
	}()

	// Simulate running for a while before stopping
	time.Sleep(30 * time.Second)
	close(doneCh)

	// Give some time for the polling to stop and the channel to close
	time.Sleep(5 * time.Second)
}
*/
