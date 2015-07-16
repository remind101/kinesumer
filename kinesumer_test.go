package kinesumer

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeTestKinesumer(t *testing.T) (*Kinesumer, *KinesisAPIMock, *ShardStateSyncMock) {
	kin := new(KinesisAPIMock)
	sssm := new(ShardStateSyncMock)
	k, err := NewKinesumer(
		kin,
		sssm,
		"TestStream",
		&DefaultKinesumerOptions,
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
	k.Stream = aws.String("c")
	kin.On("ListStreamsPages", mock.Anything, mock.Anything).Return(nil)
	e, err := k.StreamExists()
	assert.Nil(t, err)
	kin.AssertNumberOfCalls(t, "ListStreamsPages", 1)
	assert.True(t, e)
}

func TestKinesumerGetShards(t *testing.T) {
	k, kin, _ := makeTestKinesumer(t)
	k.Stream = aws.String("c")
	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(nil)
	shards, err := k.GetShards()
	assert.Nil(t, err)
	kin.AssertNumberOfCalls(t, "DescribeStreamPages", 1)
	assert.Equal(t, 2, len(shards))
	assert.Equal(t, "shard1", *shards[1].ShardID)
}

func TestKinesumerBeginEnd(t *testing.T) {
	k, kin, sssm := makeTestKinesumer(t)
	k.Stream = aws.String("c")
	kin.On("ListStreamsPages", mock.Anything, mock.Anything).Return(awserr.Error(nil))
	sssm.On("Begin", mock.Anything).Return(errors.New("bad shard sync")).Once()
	sssm.On("Begin", mock.Anything).Return(nil)
	err := k.Begin()
	assert.Error(t, err)

	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(awserr.New("bad", "bad", nil)).Once()
	err = k.Begin()
	assert.Error(t, err)

	kin.On("DescribeStreamPages", mock.Anything, mock.Anything).Return(awserr.Error(nil))
	sssm.On("GetStartSequence", mock.Anything).Return(aws.String("0")).Once()
	sssm.On("GetStartSequence", mock.Anything).Return(nil)
	kin.On("GetShardIterator", mock.Anything).Return(&kinesis.GetShardIteratorOutput{
		ShardIterator: aws.String("0"),
	}, awserr.Error(nil))
	kin.On("GetRecords", mock.Anything).Return(&kinesis.GetRecordsOutput{
		MillisBehindLatest: aws.Long(0),
		NextShardIterator:  aws.String("AAAAA"),
		Records:            []*kinesis.Record{},
	}, awserr.Error(nil))
	sssm.On("End").Return()
	assert.Nil(t, k.Begin())
	assert.Equal(t, 2, k.nRunning)
	k.End()
}
