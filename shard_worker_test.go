package kinesumer

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeTestShardWorker() (*ShardWorker, *KinesisAPIMock, *ShardStateSyncMock, <-chan Unit,
	chan<- Unit, chan *KinesisRecord) {
	kin := new(KinesisAPIMock)
	sssm := new(ShardStateSyncMock)
	buf := new(bytes.Buffer)
	stream := "TestStream"
	seq := "123"
	stop := make(chan Unit)
	stopped := make(chan Unit)
	c := make(chan *KinesisRecord)
	_0 := "0"
	_7f := "7f"
	_100 := "100"
	shard0 := "shard0"

	return &ShardWorker{
		kinesis: kin,
		logger:  logger.New(log.New(buf, "", 0)),
		shard: &kinesis.Shard{
			AdjacentParentShardID: nil,
			HashKeyRange: &kinesis.HashKeyRange{
				StartingHashKey: &_0,
				EndingHashKey:   &_7f,
			},
			ParentShardID: nil,
			SequenceNumberRange: &kinesis.SequenceNumberRange{
				StartingSequenceNumber: &_0,
				EndingSequenceNumber:   &_100,
			},
			ShardID: &shard0,
		},
		stateSync:       sssm,
		stream:          &stream,
		sequence:        &seq,
		stop:            stop,
		stopped:         stopped,
		c:               c,
		GetRecordsLimit: 123,
	}, kin, sssm, stop, stopped, c
}

func TestShardWorkerGetShardIterator(t *testing.T) {
	s, kin, _, _, _, _ := makeTestShardWorker()
	var awsNoErr awserr.Error = nil

	seq := "1234"
	iter := "AAAAA"
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: &iter,
	}, awsNoErr)
	res, err := s.GetShardIterator("TYPE", &seq)
	assert.Nil(t, err)
	assert.Equal(t, "AAAAA", *res)
}

func TestShardWorkerTryGetShardIterator(t *testing.T) {
	s, kin, _, _, _, _ := makeTestShardWorker()

	seq := "1234"
	kin.On("GetShardIterator", mock.Anything).Return(nil, awserr.New("bad", "bad", errors.New("bad")))
	res := s.TryGetShardIterator("TYPE", &seq)
	assert.Nil(t, res)
}

func TestShardWorkerGetRecords(t *testing.T) {
	s, kin, _, _, _, _ := makeTestShardWorker()
	var awsNoErr awserr.Error = nil

	var _0 int64 = 0
	iter := "AAAA"
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &iter,
		Records:            []*kinesis.Record{},
	}, awsNoErr)

	records, nextIt, mills, err := s.GetRecords(&iter)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(records))
	assert.Equal(t, iter, *nextIt)
	assert.Equal(t, _0, mills)
}

func TestShardWorkerGetRecordsAndProcess(t *testing.T) {
}

func TestShardWorkerRun(t *testing.T) {
}
