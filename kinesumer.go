package kinesumer

import (
	"errors"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/kinesumer/checkpointers/empty"
	k "github.com/remind101/kinesumer/interface"
	"github.com/remind101/kinesumer/provisioners/empty"
)

type Kinesumer struct {
	Kinesis      k.Kinesis
	Checkpointer k.Checkpointer
	Provisioner  k.Provisioner
	Stream       *string
	Options      *KinesumerOptions
	records      chan k.Record
	stop         chan Unit
	stopped      chan Unit
	nRunning     int
	rand         *rand.Rand
}

type KinesumerOptions struct {
	ListStreamsLimit    int64
	DescribeStreamLimit int64
	GetRecordsLimit     int64
	PollTime            int
	MaxShardWorkers     int
	Handlers            k.KinesumerHandlers
}

var DefaultKinesumerOptions = KinesumerOptions{
	// These values are the hard limits set by Amazon

	ListStreamsLimit:    1000,
	DescribeStreamLimit: 10000,
	GetRecordsLimit:     10000,

	PollTime:        2000,
	MaxShardWorkers: 50,
	Handlers:        k.DefaultKinesumerHandlers{},
}

func NewDefaultKinesumer(awsAccessKey, awsSecretKey, awsRegion, stream string) (*Kinesumer, error) {
	return NewKinesumer(
		kinesis.New(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
				Region:      awsRegion,
			},
		),
		nil,
		nil,
		nil,
		stream,
		nil,
	)
}

func NewKinesumer(kinesis k.Kinesis, checkpointer k.Checkpointer, provisioner k.Provisioner,
	randSource rand.Source, stream string, opt *KinesumerOptions) (*Kinesumer, error) {

	if kinesis == nil {
		return nil, k.NewKinesumerError(k.KinesumerECrit, "Kinesis object must not be nil", nil)
	}

	if checkpointer == nil {
		checkpointer = emptycheckpointer.Checkpointer{}
	}

	if provisioner == nil {
		provisioner = emptyprovisioner.Provisioner{}
	}

	if randSource == nil {
		randSource = rand.NewSource(time.Now().UnixNano())
	}

	if len(stream) == 0 {
		return nil, k.NewKinesumerError(k.KinesumerECrit, "Stream name can't be empty", nil)
	}

	if opt == nil {
		tmp := DefaultKinesumerOptions
		opt = &tmp
	}

	return &Kinesumer{
		Kinesis:      kinesis,
		Checkpointer: checkpointer,
		Provisioner:  provisioner,
		Stream:       &stream,
		Options:      opt,
		records:      make(chan k.Record, opt.GetRecordsLimit*2+10),
		rand:         rand.New(randSource),
	}, nil
}

func (kin *Kinesumer) GetStreams() (streams []*string, err error) {
	streams = make([]*string, 0)
	err = kin.Kinesis.ListStreamsPages(&kinesis.ListStreamsInput{
		Limit: &kin.Options.ListStreamsLimit,
	}, func(sts *kinesis.ListStreamsOutput, _ bool) bool {
		streams = append(streams, sts.StreamNames...)
		return true
	})
	return
}

func (kin *Kinesumer) StreamExists() (found bool, err error) {
	streams, err := kin.GetStreams()
	if err != nil {
		return
	}
	for _, stream := range streams {
		if *stream == *kin.Stream {
			return true, nil
		}
	}
	return
}

func (kin *Kinesumer) GetShards() (shards []*kinesis.Shard, err error) {
	shards = make([]*kinesis.Shard, 0)
	err = kin.Kinesis.DescribeStreamPages(&kinesis.DescribeStreamInput{
		Limit:      &kin.Options.DescribeStreamLimit,
		StreamName: kin.Stream,
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

func (kin *Kinesumer) LaunchShardWorker(shards []*kinesis.Shard) (int, error) {
	perm := kin.rand.Perm(len(shards))
	for _, j := range perm {
		err := kin.Provisioner.TryAcquire(shards[j].ShardID)
		if err == nil {
			worker := &ShardWorker{
				kinesis:         kin.Kinesis,
				shard:           shards[j],
				checkpointer:    kin.Checkpointer,
				stream:          kin.Stream,
				pollTime:        kin.Options.PollTime,
				stop:            kin.stop,
				stopped:         kin.stopped,
				c:               kin.records,
				provisioner:     kin.Provisioner,
				handlers:        kin.Options.Handlers,
				GetRecordsLimit: kin.Options.GetRecordsLimit,
			}
			kin.Options.Handlers.Go(func() {
				worker.RunWorker()
			})
			kin.nRunning++
			return j, nil
		}
	}
	return 0, errors.New("Could not launch worker")
}

func (kin *Kinesumer) Begin() (err error) {
	shards, err := kin.GetShards()
	if err != nil {
		return
	}

	err = kin.Checkpointer.Begin(kin.Options.Handlers)
	if err != nil {
		return
	}

	n := kin.Options.MaxShardWorkers
	if n <= 0 || len(shards) < n {
		n = len(shards)
	}

	start := time.Now()
	tryTime := 2 * kin.Provisioner.TTL()

	kin.stop = make(chan Unit, n)
	kin.stopped = make(chan Unit, n)
	for kin.nRunning < n && len(shards) > 0 && time.Now().Sub(start) < tryTime {
		for i := 0; i < n; i++ {
			j, err := kin.LaunchShardWorker(shards)
			if err != nil {
				kin.Options.Handlers.Err(k.NewKinesumerError(k.KinesumerEWarn, "Could not start shard worker", err))
			} else {
				shards = append(shards[:j], shards[j+1:]...)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return
}

func (kin *Kinesumer) End() {
	for kin.nRunning > 0 {
		select {
		case <-kin.stopped:
			kin.nRunning--
		case kin.stop <- Unit{}:
		}
	}
	kin.Checkpointer.End()
}

func (kin *Kinesumer) Records() <-chan k.Record {
	return kin.records
}
