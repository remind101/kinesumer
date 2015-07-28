package kinesumer

import (
	"errors"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// Unit has only one possible value, Unit{}, and is used to make signal channels to tell the workers
// when to stop
type Unit struct{}

type KinesumerIface interface {
	Begin() (err error)
	End()
	Records() <-chan *KinesisRecord
}

type Kinesumer struct {
	Kinesis      Kinesis
	Checkpointer Checkpointer
	Stream       *string
	opt          *KinesumerOptions
	records      chan *KinesisRecord
	stop         chan Unit
	stopped      chan Unit
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

func NewDefaultKinesumer(awsAccessKey, awsSecretKey, awsRegion, stream string) (*Kinesumer, error) {
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

func NewDefaultRedisKinesumer(awsAccessKey, awsSecretKey, awsRegion, redisURL, stream string) (*Kinesumer, error) {
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

func NewKinesumer(kinesis Kinesis, checkpointer Checkpointer, stream string, opt *KinesumerOptions) (*Kinesumer, error) {
	return &Kinesumer{
		Kinesis:      kinesis,
		Checkpointer: checkpointer,
		Stream:       &stream,
		opt:          opt,
		records:      make(chan *KinesisRecord, opt.GetRecordsLimit*2+10),
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

func (k *Kinesumer) LaunchShardWorker(shards []*kinesis.Shard) (int, error) {
	perm := rand.Perm(len(shards))
	for _, j := range perm {
		err := k.Checkpointer.TryAcquire(shards[j].ShardID)
		if err == nil {
			worker := &ShardWorker{
				kinesis:         k.Kinesis,
				shard:           shards[j],
				checkpointer:    k.Checkpointer,
				stream:          k.Stream,
				pollTime:        k.opt.PollTime,
				stop:            k.stop,
				stopped:         k.stopped,
				c:               k.records,
				GetRecordsLimit: k.opt.GetRecordsLimit,
			}
			go worker.RunWorker()
			k.nRunning++
			return j, nil
		}
	}
	return 0, errors.New("Could not launch worker")
}

func (k *Kinesumer) Begin() (err error) {
	shards, err := k.GetShards()
	if err != nil {
		return
	}

	err = k.Checkpointer.Begin(k.records)
	if err != nil {
		return
	}

	n := k.opt.MaxShards
	if n <= 0 || len(shards) < n {
		n = len(shards)
	}

	k.stop = make(chan Unit, k.nRunning)
	k.stopped = make(chan Unit, k.nRunning)
	for i := 0; i < n; i++ {
		j, err := k.LaunchShardWorker(shards)
		if err != nil {
			k.End()
			return err
		}
		shards = append(shards[:j], shards[j+1:]...)
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
	k.Checkpointer.End()
}

func (k *Kinesumer) Records() <-chan *KinesisRecord {
	return k.records
}
