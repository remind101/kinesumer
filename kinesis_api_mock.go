package kinesumer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/mock"
)

type KinesisAPIMock struct {
	mock.Mock
}

func (m *KinesisAPIMock) AddTagsToStream(_a0 *kinesis.AddTagsToStreamInput) (*kinesis.AddTagsToStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.AddTagsToStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.AddTagsToStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) CreateStream(_a0 *kinesis.CreateStreamInput) (*kinesis.CreateStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.CreateStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.CreateStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) DeleteStream(_a0 *kinesis.DeleteStreamInput) (*kinesis.DeleteStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.DeleteStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.DeleteStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) DescribeStream(_a0 *kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.DescribeStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.DescribeStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) DescribeStreamPages(_a0 *kinesis.DescribeStreamInput, _a1 func(*kinesis.DescribeStreamOutput, bool) bool) error {
	ret := m.Called(_a0, _a1)

	a := kinesis.Shard{
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
	}
	b := kinesis.Shard{
		AdjacentParentShardID: nil,
		HashKeyRange: &kinesis.HashKeyRange{
			StartingHashKey: aws.String("80"),
			EndingHashKey:   aws.String("ff"),
		},
		ParentShardID: nil,
		SequenceNumberRange: &kinesis.SequenceNumberRange{
			StartingSequenceNumber: aws.String("101"),
			EndingSequenceNumber:   aws.String("200"),
		},
		ShardID: aws.String("shard1"),
	}
	cont := _a1(
		&kinesis.DescribeStreamOutput{
			StreamDescription: &kinesis.StreamDescription{
				HasMoreShards: aws.Boolean(true),
				Shards:        []*kinesis.Shard{&a},
				StreamName:    aws.String("TestStream"),
				StreamStatus:  aws.String("ACTIVE"),
			},
		}, true)
	if cont {
		_a1(
			&kinesis.DescribeStreamOutput{
				StreamDescription: &kinesis.StreamDescription{
					HasMoreShards: aws.Boolean(true),
					Shards:        []*kinesis.Shard{&b},
					StreamName:    aws.String("TestStream"),
					StreamStatus:  aws.String("ACTIVE"),
				},
			}, false)
	}
	r0 := ret.Error(0)

	return r0
}
func (m *KinesisAPIMock) GetRecords(_a0 *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.GetRecordsOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.GetRecordsOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) GetShardIterator(_a0 *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.GetShardIteratorOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.GetShardIteratorOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) ListStreams(_a0 *kinesis.ListStreamsInput) (*kinesis.ListStreamsOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.ListStreamsOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.ListStreamsOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) ListStreamsPages(_a0 *kinesis.ListStreamsInput, _a1 func(*kinesis.ListStreamsOutput, bool) bool) error {
	ret := m.Called(_a0, _a1)

	cont := _a1(&kinesis.ListStreamsOutput{
		HasMoreStreams: aws.Boolean(true),
		StreamNames:    []*string{aws.String("a"), aws.String("b")},
	}, true)
	if cont {
		_a1(&kinesis.ListStreamsOutput{
			HasMoreStreams: aws.Boolean(false),
			StreamNames:    []*string{aws.String("c")},
		}, false)
	}
	r0 := ret.Error(0)

	return r0
}
func (m *KinesisAPIMock) ListTagsForStream(_a0 *kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.ListTagsForStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.ListTagsForStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) MergeShards(_a0 *kinesis.MergeShardsInput) (*kinesis.MergeShardsOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.MergeShardsOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.MergeShardsOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) PutRecord(_a0 *kinesis.PutRecordInput) (*kinesis.PutRecordOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.PutRecordOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.PutRecordOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) PutRecords(_a0 *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.PutRecordsOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.PutRecordsOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) RemoveTagsFromStream(_a0 *kinesis.RemoveTagsFromStreamInput) (*kinesis.RemoveTagsFromStreamOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.RemoveTagsFromStreamOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.RemoveTagsFromStreamOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
func (m *KinesisAPIMock) SplitShard(_a0 *kinesis.SplitShardInput) (*kinesis.SplitShardOutput, error) {
	ret := m.Called(_a0)

	var r0 *kinesis.SplitShardOutput
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*kinesis.SplitShardOutput)
	}
	r1 := ret.Error(1)

	return r0, r1
}
