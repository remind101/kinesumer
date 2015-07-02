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

func makeTestKinesumer(t *testing.T) (*Kinesumer, *KinesisAPIMock, *ShardStateSyncMock) {
	kin := new(KinesisAPIMock)
	sssm := new(ShardStateSyncMock)
	buf := new(bytes.Buffer)
	k, err := NewKinesumer(
		kin,
		sssm,
		logger.New(log.New(buf, "", 0)),
		"TestStream",
		DefaultKinesumerOptions,
	)
	if err != nil {
		t.Error(err)
	}
	return k, kin, sssm
}

func TestKinesumerGetStreams(t *testing.T) {
	k, kin, _ := makeTestKinesumer(t)
	kin.On("ListStreamsPages", mock.Anything, mock.Anything).Return(nil)
	streams, err := k.GetStreams()
	assert.Nil(t, err)
	kin.AssertNumberOfCalls(t, "ListStreamsPages", 1)
	assert.Equal(t, 3, len(streams))
	assert.Equal(t, *streams[2], "c")
}

func TestKinesumerStreamExists(t *testing.T) {
	k, kin, _ := makeTestKinesumer(t)
	stream := "c"
	k.Stream = &stream
	kin.On("ListStreamsPages", mock.Anything, mock.Anything).Return(nil)
	e, err := k.StreamExists()
	assert.Nil(t, err)
	kin.AssertNumberOfCalls(t, "ListStreamsPages", 1)
	assert.True(t, e)
}

func TestKinesumerGetShards(t *testing.T) {
	k, kin, _ := makeTestKinesumer(t)
	stream := "c"
	k.Stream = &stream
	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(nil)
	shards, err := k.GetShards()
	assert.Nil(t, err)
	kin.AssertNumberOfCalls(t, "DescribeStreamPages", 1)
	assert.Equal(t, 2, len(shards))
	assert.Equal(t, "shard1", *shards[1].ShardID)
}

func TestKinesumerBeginEnd(t *testing.T) {
	k, kin, sssm := makeTestKinesumer(t)
	awsNoErr := awserr.Error(nil)
	stream := "c"
	k.Stream = &stream
	kin.On("ListStreamsPages", mock.Anything, mock.Anything).Return(awsNoErr)
	sssm.On("Begin").Return(errors.New("bad shard sync")).Once()
	err := k.Begin()
	assert.Error(t, err)

	stream = "bad"
	sssm.On("Begin").Return(nil)
	err = k.Begin()
	assert.Error(t, err)

	stream = "c"
	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(awserr.New("bad", "bad", nil)).Once()
	err = k.Begin()
	assert.Error(t, err)

	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(awsNoErr)
	sequence := "0"
	sssm.On("GetStartSequence", mock.Anything).Return(&sequence).Once()
	sssm.On("GetStartSequence", mock.Anything).Return(nil)
	iter := "AAAAA"
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: &iter,
	}, awsNoErr)
	_0 := int64(0)
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: &_0,
		NextShardIterator:  &iter,
		Records:            []*kinesis.Record{},
	}, awsNoErr)
	sssm.On("End").Return()
	assert.Nil(t, k.Begin())
	assert.Equal(t, 2, k.nRunning)
	k.End()
}
