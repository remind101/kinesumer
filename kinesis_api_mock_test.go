package kinesumer

import "github.com/stretchr/testify/mock"

import "github.com/aws/aws-sdk-go/service/kinesis"

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

	sn := "TestStream"
	active := "ACTIVE"
	tr := true
	_0 := "0"
	_100 := "100"
	_101 := "101"
	_200 := "200"
	shard0 := "shard0"
	shard1 := "shard1"
	_7f := "7f"
	_80 := "80"
	_ff := "ff"
	a := kinesis.Shard{
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
	}
	b := kinesis.Shard{
		AdjacentParentShardID: nil,
		HashKeyRange: &kinesis.HashKeyRange{
			StartingHashKey: &_80,
			EndingHashKey:   &_ff,
		},
		ParentShardID: nil,
		SequenceNumberRange: &kinesis.SequenceNumberRange{
			StartingSequenceNumber: &_101,
			EndingSequenceNumber:   &_200,
		},
		ShardID: &shard1,
	}
	cont := _a1(
		&kinesis.DescribeStreamOutput{
			StreamDescription: &kinesis.StreamDescription{
				HasMoreShards: &tr,
				Shards:        []*kinesis.Shard{&a},
				StreamName:    &sn,
				StreamStatus:  &active,
			},
		}, true)
	if cont {
		_a1(
			&kinesis.DescribeStreamOutput{
				StreamDescription: &kinesis.StreamDescription{
					HasMoreShards: &tr,
					Shards:        []*kinesis.Shard{&b},
					StreamName:    &sn,
					StreamStatus:  &active,
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

	tr := true
	fa := false
	a := "a"
	b := "b"
	c := "c"
	cont := _a1(&kinesis.ListStreamsOutput{
		HasMoreStreams: &tr,
		StreamNames:    []*string{&a, &b},
	}, true)
	if cont {
		_a1(&kinesis.ListStreamsOutput{
			HasMoreStreams: &fa,
			StreamNames:    []*string{&c},
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
