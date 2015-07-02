package kinesumer

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/remind101/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeTestShardWorker() (*ShardWorker, *KinesisAPIMock, *ShardStateSyncMock, chan Unit,
	chan Unit, chan *KinesisRecord, *bytes.Buffer) {
	kin := new(KinesisAPIMock)
	sssm := new(ShardStateSyncMock)
	buf := new(bytes.Buffer)
	stream := "TestStream"
	seq := "123"
	stop := make(chan Unit, 1)
	stopped := make(chan Unit, 1)
	c := make(chan *KinesisRecord, 100)
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
	}, kin, sssm, stop, stopped, c, buf
}

func TestShardWorkerGetShardIterator(t *testing.T) {
	s, kin, _, _, _, _, _ := makeTestShardWorker()
	awsNoErr := awserr.Error(nil)

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
	s, kin, _, _, _, _, _ := makeTestShardWorker()

	seq := "1234"
	kin.On("GetShardIterator", mock.Anything).Return(nil, awserr.New("bad", "bad", errors.New("bad")))
	assert.Panics(t, func() {
		s.TryGetShardIterator("TYPE", &seq)
	})
}

func TestShardWorkerGetRecords(t *testing.T) {
	s, kin, _, _, _, _, _ := makeTestShardWorker()
	awsNoErr := awserr.Error(nil)

	_0 := int64(0)
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
	s, kin, sssm, stp, _, c, buf := makeTestShardWorker()

	_0 := int64(0)
	awsNoErr := awserr.Error(nil)
	it := "AAAA"
	seq := "123"
	partKey := "aaaa"
	record1 := kinesis.Record{
		Data:           []byte("help I'm trapped"),
		PartitionKey:   &partKey,
		SequenceNumber: &seq,
	}
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{&record1},
	}, awsNoErr).Once()
	doneC := make(chan *KinesisRecord)
	sssm.On("DoneC").Return(doneC)
	brk, nextIt, nextSeq := s.GetRecordsAndProcess(&it, &seq)
	rec := <-c
	assert.Equal(t, &record1, rec.Record)
	assert.False(t, brk)
	assert.Equal(t, it, *nextIt)
	assert.Equal(t, seq, *nextSeq)

	stp <- Unit{}
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{},
	}, awserr.New("bad", "bad", nil))
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: &it,
	}, awsNoErr)
	brk, nextIt, nextSeq = s.GetRecordsAndProcess(&it, &seq)
	assert.True(t, strings.Contains(buf.String(), "GetRecords encountered an error"))
	kin.AssertNumberOfCalls(t, "GetShardIterator", 1)
	assert.True(t, brk)
}

func TestShardWorkerRun(t *testing.T) {
	s, kin, sssm, stp, stpd, c, _ := makeTestShardWorker()
	sssm.On("GetStartSequence", mock.Anything).Return(nil)

	awsNoErr := awserr.Error(nil)
	_0 := int64(0)
	it := "AAAA"
	seq := "123"
	partKey := "aaaa"
	record1 := kinesis.Record{
		Data:           []byte("help I'm trapped"),
		PartitionKey:   &partKey,
		SequenceNumber: &seq,
	}
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{&record1},
	}, awsNoErr).Once()
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{},
	}, awsNoErr)
	doneC := make(chan *KinesisRecord)
	sssm.On("DoneC").Return(doneC)
	iter := "AAAA"
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: &iter,
	}, awsNoErr)
	go func() {
		time.Sleep(10 * time.Millisecond)
		stp <- Unit{}
	}()
	s.RunWorker()
	<-stpd
	rec := <-c
	assert.Equal(t, &record1, rec.Record)
}
