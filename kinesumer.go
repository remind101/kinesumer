package kinesumer

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// Unit has only one possible value, Unit{}, and is used to make signal channels to tell the workers
// when to stop
type Unit struct{}

type KinesumerAPI interface {
	Begin() (err error)
	End()
	Records() <-chan *KinesisRecord
}

type Kinesumer struct {
	Kinesis   KinesisAPI
	StateSync ShardStateSync
	Stream    *string
	opt       *KinesumerOptions
	records   chan *KinesisRecord
	stop      chan Unit
	stopped   chan Unit
	nRunning  int
}

type KinesumerOptions struct {
	ListStreamsLimit    int64
	DescribeStreamLimit int64
	GetRecordsLimit     int64
	PollTime            int
}

var DefaultKinesumerOptions = KinesumerOptions{
	ListStreamsLimit:    1000,
	DescribeStreamLimit: 10000,
	GetRecordsLimit:     50,
	PollTime:            2000,
}

func NewDefaultKinesumer(awsAccessKey, awsSecretKey, awsRegion, stream string) (*Kinesumer, error) {
	return NewKinesumer(
		kinesis.New(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
				Region:      awsRegion,
			},
		),
		&EmptyStateSync{},
		stream,
		&DefaultKinesumerOptions)
}

func NewDefaultRedisKinesumer(awsAccessKey, awsSecretKey, awsRegion, redisURL, stream string) (*Kinesumer, error) {
	rss, err := NewRedisStateSync(&RedisStateSyncOptions{
		Ticker: time.NewTicker(5 * time.Second).C,
	})
	if err != nil {
		return nil, err
	}
	return NewKinesumer(
		kinesis.New(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
				Region:      awsRegion,
			},
		),
		rss,
		stream,
		&DefaultKinesumerOptions)
}

func NewKinesumer(kinesis KinesisAPI, stateSync ShardStateSync, stream string, opt *KinesumerOptions) (*Kinesumer, error) {
	return &Kinesumer{
		Kinesis:   kinesis,
		StateSync: stateSync,
		Stream:    &stream,
		opt:       opt,
		records:   make(chan *KinesisRecord, opt.GetRecordsLimit*2+10),
	}, nil
}

func (k *Kinesumer) GetStreams() (streams []*string, err error) {
	streams = make([]*string, 0)
	err = k.Kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &k.opt.ListStreamsLimit,
	}, func(sts *kinesis.ListStreamsOutput, _ bool) bool {
		streams = append(streams, sts.StreamNames...)
		return true
	})
	return
}

func (k *Kinesumer) StreamExists() (found bool, err error) {
	streams, err := k.GetStreams()
	if err != nil {
		return
	}
	for _, stream := range streams {
		if *stream == *k.Stream {
			return true, nil
		}
	}
	return
}

func (k *Kinesumer) GetShards() (shards []*kinesis.Shard, err error) {
	shards = make([]*kinesis.Shard, 0)
	err = k.Kinesis.DescribeStreamPages(&kinesis.DescribeStreamInput{
		Limit:      &k.opt.DescribeStreamLimit,
		StreamName: k.Stream,
	}, func(desc *kinesis.DescribeStreamOutput, _ bool) bool {
		if desc == nil {
			err = errors.New("Stream could not be described")
			return false
		}
		if *desc.StreamDescription.StreamStatus == "DELETING" {
			err = errors.New("Stream is being deleted")
			return false
		}
		shards = append(shards, desc.StreamDescription.Shards...)
		return true
	})
	return
}

func (k *Kinesumer) Begin() (err error) {
	shards, err := k.GetShards()
	if err != nil {
		return
	}

	err = k.StateSync.Begin(k.records)
	if err != nil {
		return
	}

	k.nRunning = len(shards)
	k.stop = make(chan Unit, k.nRunning)
	k.stopped = make(chan Unit, k.nRunning)
	for _, shard := range shards {
		worker := &ShardWorker{
			kinesis:         k.Kinesis,
			shard:           shard,
			stateSync:       k.StateSync,
			stream:          k.Stream,
			pollTime:        k.opt.PollTime,
			stop:            k.stop,
			stopped:         k.stopped,
			c:               k.records,
			GetRecordsLimit: k.opt.GetRecordsLimit,
		}
		go worker.RunWorker()
	}

	return
}

func (k *Kinesumer) End() {
	for k.nRunning > 0 {
		select {
		case <-k.stopped:
			k.nRunning--
		case k.stop <- Unit{}:
		}
	}
	k.StateSync.End()
}

func (k *Kinesumer) Records() <-chan *KinesisRecord {
	return k.records
}
