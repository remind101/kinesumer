package kinesumeriface

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
)

type Kinesis interface {
	AddTagsToStream(*kinesis.AddTagsToStreamInput) (*kinesis.AddTagsToStreamOutput, error)

	CreateStream(*kinesis.CreateStreamInput) (*kinesis.CreateStreamOutput, error)

	DeleteStream(*kinesis.DeleteStreamInput) (*kinesis.DeleteStreamOutput, error)

	DescribeStream(*kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error)

	DescribeStreamPages(input *kinesis.DescribeStreamInput, fn func(p *kinesis.DescribeStreamOutput, lastPage bool) (shouldContinue bool)) error

	GetRecords(*kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)

	GetShardIterator(*kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)

	ListStreams(*kinesis.ListStreamsInput) (*kinesis.ListStreamsOutput, error)

	ListStreamsPages(input *kinesis.ListStreamsInput, fn func(p *kinesis.ListStreamsOutput, lastPage bool) (shouldContinue bool)) error

	ListTagsForStream(*kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error)

	MergeShards(*kinesis.MergeShardsInput) (*kinesis.MergeShardsOutput, error)

	PutRecord(*kinesis.PutRecordInput) (*kinesis.PutRecordOutput, error)

	PutRecords(*kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)

	RemoveTagsFromStream(*kinesis.RemoveTagsFromStreamInput) (*kinesis.RemoveTagsFromStreamOutput, error)

	SplitShard(*kinesis.SplitShardInput) (*kinesis.SplitShardOutput, error)
}
