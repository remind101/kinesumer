package kinesumer

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
)

type Kinesumer struct {
	opt      *KinesumerOptions
	kinesis  *kinesis.Kinesis
	C        chan *ShardState
	stop     chan struct{}
	stopped  chan struct{}
	nRunning int
}

type KinesumerOptions struct {
	Kinesis             *kinesis.Kinesis
	StateSync           ShardStateSync
	Logger              logger.Logger
	Stream              *string
	ListStreamsLimit    int64
	DescribeStreamLimit int64
	GetRecordsLimit     int64
}

func DefaultKinesumerOptions() *KinesumerOptions {
	return &KinesumerOptions{
		StateSync:           NewEmptyStateSync(),
		Logger:              logger.New(log.New(os.Stdout, "", 0)),
		ListStreamsLimit:    1000,
		DescribeStreamLimit: 10000,
		GetRecordsLimit:     50,
	}
}

func NewKinesumer(opt *KinesumerOptions) (*Kinesumer, error) {
	return &Kinesumer{
		opt:     opt,
		kinesis: opt.Kinesis,
		C:       make(chan *ShardState, opt.GetRecordsLimit),
	}, nil
}

func (k *Kinesumer) GetStreams() (streams []*string) {
	streams = make([]*string, 0)
	k.kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &k.opt.ListStreamsLimit,
	}, func(sts *kinesis.ListStreamsOutput, _ bool) bool {
		streams = append(streams, sts.StreamNames...)
		return true
	})
	return
}

func (k *Kinesumer) StreamExists() (found bool) {
	k.kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &k.opt.ListStreamsLimit,
	}, func(streams *kinesis.ListStreamsOutput, _ bool) bool {
		for _, stream := range streams.StreamNames {
			if *stream == *k.opt.Stream {
				found = true
				return false
			}
		}
		return true
	})
	return
}

func (k *Kinesumer) GetShards() (shards []*kinesis.Shard) {
	shards = make([]*kinesis.Shard, 0)
	k.kinesis.DescribeStreamPages(&kinesis.DescribeStreamInput{
		Limit:      &k.opt.DescribeStreamLimit,
		StreamName: k.opt.Stream,
	}, func(desc *kinesis.DescribeStreamOutput, _ bool) bool {
		if *desc.StreamDescription.StreamStatus == "DELETING" {
			k.opt.Logger.Crit("Stream is being deleted", "stream", *k.opt.Stream)
			panic("Stream is being deleted")
		}
		shards = append(shards, desc.StreamDescription.Shards...)
		return true
	})
	return
}

func (k *Kinesumer) Begin() {
	if !k.StreamExists() {
		k.opt.Logger.Crit("Stream not found", "stream", *k.opt.Stream)
		panic("Stream not found")
	}

	k.opt.StateSync.begin()

	shards := k.GetShards()
	k.nRunning = len(shards)
	k.stop = make(chan struct{}, k.nRunning)
	k.stopped = make(chan struct{}, k.nRunning)
	for _, shard := range shards {
		worker := &ShardWorker{
			kinesis:         k.kinesis,
			logger:          k.opt.Logger,
			shard:           shard,
			stateSync:       k.opt.StateSync,
			stream:          k.opt.Stream,
			stop:            &k.stop,
			stopped:         &k.stopped,
			c:               &k.C,
			GetRecordsLimit: k.opt.GetRecordsLimit,
		}
		go worker.RunWorker()
	}
}

func (k *Kinesumer) End() {
	k.opt.StateSync.end()

	for k.nRunning > 0 {
		select {
		case <-k.stopped:
			k.nRunning--
		case k.stop <- struct{}{}:
		}
	}
}
