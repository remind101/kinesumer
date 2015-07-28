package kinesumer

import (
	"errors"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
	k "github.com/remind101/kinesumer/interface"
)

type KinesumerImpl struct {
	Kinesis      k.Kinesis
	Checkpointer k.Checkpointer
	Stream       *string
	opt          *KinesumerOptions
	records      chan *k.KinesisRecord
	stop         chan k.Unit
	stopped      chan k.Unit
	nRunning     int
}

type KinesumerOptions struct {
	ListStreamsLimit    int64
	DescribeStreamLimit int64
	GetRecordsLimit     int64
	PollTime            int
	MaxShards           int
}

var DefaultKinesumerOptions = KinesumerOptions{
	ListStreamsLimit:    1000,
	DescribeStreamLimit: 10000,
	GetRecordsLimit:     50,
	PollTime:            2000,
}

func NewDefaultKinesumer(awsAccessKey, awsSecretKey, awsRegion, stream string) (*KinesumerImpl, error) {
	return NewKinesumer(
		kinesis.New(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
				Region:      awsRegion,
			},
		),
		&EmptyCheckpointer{},
		stream,
		&DefaultKinesumerOptions)
}

func NewDefaultRedisKinesumer(awsAccessKey, awsSecretKey, awsRegion, redisURL, stream string) (*KinesumerImpl, error) {
	rss, err := NewRedisCheckpointer(&RedisCheckpointerOptions{
		SavePeriod: 5 * time.Second,
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

func NewKinesumer(kinesis k.Kinesis, checkpointer k.Checkpointer, stream string, opt *KinesumerOptions) (*KinesumerImpl, error) {
	return &KinesumerImpl{
		Kinesis:      kinesis,
		Checkpointer: checkpointer,
		Stream:       &stream,
		opt:          opt,
		records:      make(chan *k.KinesisRecord, opt.GetRecordsLimit*2+10),
	}, nil
}

func (ki *KinesumerImpl) GetStreams() (streams []*string, err error) {
	streams = make([]*string, 0)
	err = ki.Kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &ki.opt.ListStreamsLimit,
	}, func(sts *kinesis.ListStreamsOutput, _ bool) bool {
		streams = append(streams, sts.StreamNames...)
		return true
	})
	return
}

func (ki *KinesumerImpl) StreamExists() (found bool, err error) {
	streams, err := ki.GetStreams()
	if err != nil {
		return
	}
	for _, stream := range streams {
		if *stream == *ki.Stream {
			return true, nil
		}
	}
	return
}

func (ki *KinesumerImpl) GetShards() (shards []*kinesis.Shard, err error) {
	shards = make([]*kinesis.Shard, 0)
	err = ki.Kinesis.DescribeStreamPages(&kinesis.DescribeStreamInput{
		Limit:      &ki.opt.DescribeStreamLimit,
		StreamName: ki.Stream,
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

func (ki *KinesumerImpl) LaunchShardWorker(shards []*kinesis.Shard) (int, error) {
	perm := rand.Perm(len(shards))
	for _, j := range perm {
		err := ki.Checkpointer.TryAcquire(shards[j].ShardID)
		if err == nil {
			worker := &ShardWorker{
				kinesis:         ki.Kinesis,
				shard:           shards[j],
				checkpointer:    ki.Checkpointer,
				stream:          ki.Stream,
				pollTime:        ki.opt.PollTime,
				stop:            ki.stop,
				stopped:         ki.stopped,
				c:               ki.records,
				GetRecordsLimit: ki.opt.GetRecordsLimit,
			}
			go worker.RunWorker()
			ki.nRunning++
			return j, nil
		}
	}
	return 0, errors.New("Could not launch worker")
}

func (ki *KinesumerImpl) Begin() (err error) {
	shards, err := ki.GetShards()
	if err != nil {
		return
	}

	err = ki.Checkpointer.Begin(ki.records)
	if err != nil {
		return
	}

	n := ki.opt.MaxShards
	if n <= 0 || len(shards) < n {
		n = len(shards)
	}

	ki.stop = make(chan k.Unit, ki.nRunning)
	ki.stopped = make(chan k.Unit, ki.nRunning)
	for i := 0; i < n; i++ {
		j, err := ki.LaunchShardWorker(shards)
		if err != nil {
			ki.End()
			return err
		}
		shards = append(shards[:j], shards[j+1:]...)
	}

	return
}

func (ki *KinesumerImpl) End() {
	for ki.nRunning > 0 {
		select {
		case <-ki.stopped:
			ki.nRunning--
		case ki.stop <- k.Unit{}:
		}
	}
	ki.Checkpointer.End()
}

func (ki *KinesumerImpl) Records() <-chan *k.KinesisRecord {
	return ki.records
}
