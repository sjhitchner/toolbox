package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sjhitchner/toolbox/pkg/streaming"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	zl "github.com/rs/zerolog"
)

type ShardIteratorType string

const (
	// The following are the valid Amazon Kinesis shard iterator types:
	//
	//    * AT_SEQUENCE_NUMBER - Start reading from the position denoted by a specific
	//    sequence number, provided in the value StartingSequenceNumber.
	AtSequenceNumber ShardIteratorType = kinesis.ShardIteratorTypeAtSequenceNumber

	//
	//    * AFTER_SEQUENCE_NUMBER - Start reading right after the position denoted
	//    by a specific sequence number, provided in the value StartingSequenceNumber.
	AfterSequenceNumber ShardIteratorType = kinesis.ShardIteratorTypeAfterSequenceNumber
	//
	//    * AT_TIMESTAMP - Start reading from the position denoted by a specific
	//    time stamp, provided in the value Timestamp.
	AtTimestamp ShardIteratorType = kinesis.ShardIteratorTypeAtTimestamp

	//
	//    * TRIM_HORIZON - Start reading at the last untrimmed record in the shard
	//    in the system, which is the oldest data record in the shard.
	TrimHorizon ShardIteratorType = kinesis.ShardIteratorTypeTrimHorizon

	//
	//    * LATEST - Start reading just after the most recent record in the shard,
	//    so that you always read the most recent data in the shard.
	Latest ShardIteratorType = kinesis.ShardIteratorTypeLatest
)

type Kinesis[T any] struct {
	sess       *session.Session
	srv        *kinesis.Kinesis
	streamName string
	done       chan struct{}
	logger     zl.Logger
}

func New[T any](sess *session.Session, streamName string, logger *zl.Logger) *Kinesis[T] {

	srv := kinesis.New(sess)

	log := logger.With().
		Str("stream", streamName).
		Logger()

	return &Kinesis[T]{
		sess:       sess,
		srv:        srv,
		streamName: streamName,
		done:       make(chan struct{}),
		logger:     log,
	}
}

func (t *Kinesis[T]) Consume() (<-chan T, error) {
	// Get Shard Info
	shards, err := t.srv.DescribeStream(
		&kinesis.DescribeStreamInput{
			StreamName: aws.String(t.streamName),
		})
	if err != nil {
		return nil, err
	}

	fmt.Println(shards)

	shardStreams := make([]<-chan T, 0, len(shards.StreamDescription.Shards))
	for _, shard := range shards.StreamDescription.Shards {
		ch, err := t.consumeShard(
			Latest,
			aws.StringValue(shard.ShardId),
			aws.StringValue(shard.SequenceNumberRange.StartingSequenceNumber),
			time.Now().UTC(),
		)
		if err != nil {
			return nil, err
		}

		shardStreams = append(shardStreams, ch)
	}

	out := streaming.Merge[T](t.done, shardStreams...)

	return out, nil
}

func (t *Kinesis[T]) consumeShard(iteratorType ShardIteratorType, shardID, sequenceNumber string, timestamp time.Time) (<-chan T, error) {

	var input *kinesis.GetShardIteratorInput

	switch iteratorType {
	case AtSequenceNumber, AfterSequenceNumber:
		input = &kinesis.GetShardIteratorInput{
			StreamName:             aws.String(t.streamName),
			ShardId:                aws.String(shardID),
			ShardIteratorType:      aws.String(string(iteratorType)),
			StartingSequenceNumber: aws.String(sequenceNumber),
		}
	case AtTimestamp, TrimHorizon:
		input = &kinesis.GetShardIteratorInput{
			StreamName:        aws.String(t.streamName),
			ShardId:           aws.String(shardID),
			ShardIteratorType: aws.String(string(iteratorType)),
			Timestamp:         aws.Time(timestamp),
		}
	default:
		input = &kinesis.GetShardIteratorInput{
			StreamName:        aws.String(t.streamName),
			ShardId:           aws.String(shardID),
			ShardIteratorType: aws.String("LATEST"),
		}
	}

	iteratorOutput, err := t.srv.GetShardIterator(input)
	if err != nil {
		return nil, err
	}

	logger := t.logger.With().
		Str("shard_id", shardID).
		Logger()

	out := make(chan []byte)

	go func() {
		defer close(out)

		shardIterator := iteratorOutput.ShardIterator

		for {
			// get records use shard iterator for making request
			records, err := t.srv.GetRecords(&kinesis.GetRecordsInput{
				Limit:         aws.Int64(10000),
				ShardIterator: shardIterator,
			})
			if err != nil {
				if aws.StringValue(records.NextShardIterator) == "" || shardIterator == records.NextShardIterator {
					logger.Error().
						Err(err).
						Msg("error retrieving from shard")
				}

				logger.Error().
					Err(err).
					Msg("waiting for more messages")
				<-time.After(3 * time.Second)
				continue
			}

			// process the data
			if len(records.Records) > 0 {
				for _, d := range records.Records {
					out <- d.Data
				}
			}

			shardIterator = records.NextShardIterator
			<-time.After(1 * time.Second)
		}
	}()

	morphed := streaming.Morph[[]byte, T](t.done, out, func(b []byte) (T, error) {
		var obj T

		if err := json.Unmarshal(b, &obj); err != nil {
			t.logger.Error().
				Err(err).
				Str("data", string(b)).
				Msg("failed deserializing object")
			return obj, err
		}

		return obj, nil
	})

	return morphed, nil
}

func (t *Kinesis[T]) Publish(in <-chan T) error {

	ctx := context.Background()

	morphed := streaming.Morph[T, []byte](t.done, in, func(obj T) ([]byte, error) {
		b, err := json.Marshal(obj)
		if err != nil {
			t.logger.Error().
				Err(err).
				Interface("obj", obj).
				Msg("failed serialing object")
			return b, err
		}

		return b, nil
	})

	for data := range morphed {
		if data == nil {
			continue
		}

		input := &kinesis.PutRecordInput{
			StreamName:   aws.String(t.streamName),
			Data:         data,
			PartitionKey: aws.String("1"),
			//ExplicitHashKey: aws.String(),
			// SequenceNumberForOrdering *string `type:"string"`
		}

		output, err := t.srv.PutRecordWithContext(ctx, input)
		if err != nil {
			t.logger.Error().
				Err(err).
				Str("data", string(data)).
				Msg("failed publishing message")
			continue
		}

		t.logger.Info().
			Str("data", string(data)).
			Str("sequence_id", aws.StringValue(output.SequenceNumber)).
			Str("shard_id", aws.StringValue(output.ShardId)).
			Msg("published message")
	}

	return nil
}

func (t *Kinesis[T]) PublishBatch(in <-chan T, batchSize int, timeout time.Duration) error {

	morphed := streaming.Morph[T, []byte](t.done, in, func(obj T) ([]byte, error) {
		b, err := json.Marshal(obj)
		if err != nil {
			t.logger.Error().
				Err(err).
				Interface("obj", obj).
				Msg("failed serialing object")
			return b, err
		}

		return b, nil
	})

	batchedStream := batchRecords(morphed, batchSize, timeout)
	retryStream := make(chan []*kinesis.PutRecordsRequestEntry)
	ch := streaming.Merge[[]*kinesis.PutRecordsRequestEntry](t.done, batchedStream, retryStream)

	return t.publishBatch(ch, retryStream)
}

func (t *Kinesis[T]) publishBatch(in <-chan []*kinesis.PutRecordsRequestEntry, retry chan<- []*kinesis.PutRecordsRequestEntry) error {
	var count int64
	for records := range in {
		input := &kinesis.PutRecordsInput{
			StreamName: aws.String(t.streamName),
			Records:    records,
		}

		output, err := t.srv.PutRecords(input)
		if err != nil {
			t.logger.Error().
				Err(err).
				Int("batch", len(records)).
				Msg("failed publishing batch")
			continue
		}

		failedRecords := aws.Int64Value(output.FailedRecordCount)

		if failedRecords > 0 {
			retry <- records
			<-time.After(200 * time.Millisecond)
		} else {
			count += int64(len(output.Records))
		}

		t.logger.Info().
			Int64("failed_records", failedRecords).
			Int("records", len(output.Records)).
			Int64("count", count).
			Msg("published message")
	}

	return nil
}

func batchRecords(in <-chan []byte, batchSize int, timeout time.Duration) <-chan []*kinesis.PutRecordsRequestEntry {

	out := make(chan []*kinesis.PutRecordsRequestEntry)

	go func() {
		defer close(out)

		buf := make([]*kinesis.PutRecordsRequestEntry, 0, batchSize)
		ticker := time.NewTicker(timeout)

		for i := 0; ; {
			select {
			case data, ok := <-in:
				if !ok {
					if len(buf) > 0 {
						out <- buf
					}
					return
				}

				buf = append(buf, &kinesis.PutRecordsRequestEntry{
					Data:         data,
					PartitionKey: aws.String(fmt.Sprintf("key%02d", i%12)),
					// ExplicitHashKey: aws.String(),
					// SequenceNumberForOrdering *string `type:"string"`
				})

				i++

				ticker.Reset(timeout)
				if i >= batchSize {
					if len(buf) > 0 {
						out <- buf
					}
					i = 0
					buf = make([]*kinesis.PutRecordsRequestEntry, 0, batchSize)
				}

			case <-ticker.C:
				if len(buf) > 0 {
					out <- buf
				}
				i = 0
				buf = make([]*kinesis.PutRecordsRequestEntry, 0, batchSize)
			}
		}
	}()

	return out
}
