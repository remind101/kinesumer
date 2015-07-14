package kinesumer

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	stop := make(chan Unit, 1)
	stopped := make(chan Unit, 1)
	c := make(chan *KinesisRecord, 100)

	return &ShardWorker{
		kinesis: kin,
		logger:  logger.New(log.New(buf, "", 0)),
		shard: &kinesis.Shard{
			AdjacentParentShardID: nil,
			HashKeyRange: &kinesis.HashKeyRange{
				StartingHashKey: aws.String("0"),
				EndingHashKey:   aws.String("7f"),
			},
			ParentShardID: nil,
			SequenceNumberRange: &kinesis.SequenceNumberRange{
				StartingSequenceNumber: aws.String("0"),
				EndingSequenceNumber:   aws.String("100"),
			},
			ShardID: aws.String("shard0"),
		},
		stateSync:       sssm,
		stream:          aws.String("TestStream"),
		sequence:        aws.String("123"),
		stop:            stop,
		stopped:         stopped,
		c:               c,
		GetRecordsLimit: 123,
	}, kin, sssm, stop, stopped, c, buf
}

func TestShardWorkerGetShardIterator(t *testing.T) {
	s, kin, _, _, _, _, _ := makeTestShardWorker()
	awsNoErr := awserr.Error(nil)

	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: aws.String("AAAAA"),
	}, awsNoErr)
	res, err := s.GetShardIterator("TYPE", aws.String("123"))
	assert.Nil(t, err)
	assert.Equal(t, "AAAAA", *res)
}

func TestShardWorkerTryGetShardIterator(t *testing.T) {
	s, kin, _, _, _, _, _ := makeTestShardWorker()

	kin.On("GetShardIterator", mock.Anything).Return(nil, awserr.New("bad", "bad", errors.New("bad")))
	assert.Panics(t, func() {
		s.TryGetShardIterator("TYPE", aws.String("123"))
	})
}

func TestShardWorkerGetRecords(t *testing.T) {
	s, kin, _, _, _, _, _ := makeTestShardWorker()
	awsNoErr := awserr.Error(nil)

	iter := "AAAA"
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: aws.Long(0),
		NextShardIterator:  &iter,
		Records:            []*kinesis.Record{},
	}, awsNoErr)

	records, nextIt, mills, err := s.GetRecords(&iter)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(records))
	assert.Equal(t, iter, *nextIt)
	assert.Equal(t, int64(0), mills)
}

func TestShardWorkerGetRecordsAndProcess(t *testing.T) {
	s, kin, sssm, stp, _, c, buf := makeTestShardWorker()

	awsNoErr := awserr.Error(nil)
	it := "AAAA"
	partKey := "aaaa"
	record1 := kinesis.Record{
		Data:           []byte("help I'm trapped"),
		PartitionKey:   &partKey,
		SequenceNumber: aws.String("123"),
	}
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: aws.Long(0),
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{&record1},
	}, awsNoErr).Once()
	doneC := make(chan *KinesisRecord)
	sssm.On("DoneC").Return(doneC)
	brk, nextIt, nextSeq := s.GetRecordsAndProcess(&it, aws.String("123"))
	rec := <-c
	assert.Equal(t, &record1, rec.Record)
	assert.False(t, brk)
	assert.Equal(t, it, *nextIt)
	assert.Equal(t, "123", *nextSeq)

	stp <- Unit{}
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: aws.Long(0),
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{},
	}, awserr.New("bad", "bad", nil))
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: &it,
	}, awsNoErr)
	brk, nextIt, nextSeq = s.GetRecordsAndProcess(&it, aws.String("123"))
	assert.True(t, strings.Contains(buf.String(), "GetRecords encountered an error"))
	kin.AssertNumberOfCalls(t, "GetShardIterator", 1)
	assert.True(t, brk)
}

func TestShardWorkerRun(t *testing.T) {
	s, kin, sssm, stp, stpd, c, _ := makeTestShardWorker()
	sssm.On("GetStartSequence", mock.Anything).Return(nil)

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
		MillisBehindLatest: aws.Long(0),
		NextShardIterator:  &it,
		Records:            []*kinesis.Record{&record1},
	}, awsNoErr).Once()
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: aws.Long(0),
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
