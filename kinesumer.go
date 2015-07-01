package kinesumer

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type Unit struct{}

type Kinesumer struct {
	Kinesis   *kinesis.Kinesis
	StateSync ShardStateSync
	Logger    logger.Logger
	Stream    *string
	opt       *KinesumerOptions
	Records   chan *KinesisRecord
	stop      chan Unit
	stopped   chan Unit
	nRunning  int
}

type KinesumerOptions struct {
	ListStreamsLimit    int64
	DescribeStreamLimit int64
	GetRecordsLimit     int64
}

var DefaultKinesumerOptions = KinesumerOptions{
	ListStreamsLimit:    1000,
	DescribeStreamLimit: 10000,
	GetRecordsLimit:     50,
}

func NewKinesumer(kinesis *kinesis.Kinesis, stateSync ShardStateSync, logger logger.Logger,
	stream string, opt KinesumerOptions) (*Kinesumer, error) {
	return &Kinesumer{
		Kinesis:   kinesis,
		StateSync: stateSync,
		Logger:    logger,
		Stream:    &stream,
		opt:       &opt,
		Records:   make(chan *KinesisRecord, opt.GetRecordsLimit),
	}, nil
}

func (k *Kinesumer) GetStreams() (streams []*string) {
	streams = make([]*string, 0)
	k.Kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &k.opt.ListStreamsLimit,
	}, func(sts *kinesis.ListStreamsOutput, _ bool) bool {
		streams = append(streams, sts.StreamNames...)
		return true
	})
	return
}

func (k *Kinesumer) StreamExists() (found bool) {
	for _, stream := range k.GetStreams() {
		if *stream == *k.Stream {
			return true
		}
	}
	return
}

func (k *Kinesumer) GetShards() (shards []*kinesis.Shard, err error) {
	shards = make([]*kinesis.Shard, 0)
	k.Kinesis.DescribeStreamPages(&kinesis.DescribeStreamInput{
		Limit:      &k.opt.DescribeStreamLimit,
		StreamName: k.Stream,
	}, func(desc *kinesis.DescribeStreamOutput, _ bool) bool {
		if *desc.StreamDescription.StreamStatus == "DELETING" {
			k.Logger.Crit("Stream is being deleted", "stream", *k.Stream)
			err = errors.New("Stream is being deleted")
			return false
		}
		shards = append(shards, desc.StreamDescription.Shards...)
		return true
	})
	return
}

func (k *Kinesumer) Begin() error {
	if !k.StreamExists() {
		k.Logger.Crit("Stream not found", "stream", *k.Stream)
		return errors.New("Stream not found")
	}

	err := k.StateSync.Begin()
	if err != nil {
		return err
	}

	shards, err := k.GetShards()
	if err != nil {
		return err
	}
	k.nRunning = len(shards)
	k.stop = make(chan Unit, k.nRunning)
	k.stopped = make(chan Unit, k.nRunning)
	for _, shard := range shards {
		worker := &ShardWorker{
			kinesis:         k.Kinesis,
			logger:          k.Logger,
			shard:           shard,
			stateSync:       k.StateSync,
			stream:          k.Stream,
			stop:            k.stop,
			stopped:         k.stopped,
			c:               k.Records,
			GetRecordsLimit: k.opt.GetRecordsLimit,
		}
		go worker.RunWorker()
	}

	return nil
}

func (k *Kinesumer) End() {
	k.StateSync.End()

	for k.nRunning > 0 {
		select {
		case <-k.stopped:
			k.nRunning--
		case k.stop <- Unit{}:
		}
	}
}
