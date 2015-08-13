// +build integration

package kinesumer

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pborman/uuid"
	"github.com/remind101/kinesumer/checkpointers/redis"
	"github.com/remind101/kinesumer/provisioners/redis"
	"github.com/remind101/kinesumer/redispool"
)

func consecAll(n string, char rune) bool {
	for _, c := range n {
		if c != char {
			return false
		}
	}
	return true
}

func consec(n, sn string) bool {
	if len(n) == 0 || len(sn) == 0 {
		return false
	} else if n[0] == sn[0] {
		return consec(n[1:], sn[1:])
	} else if n[0]+1 == sn[0] {
		return consecAll(n[1:], '9') && consecAll(sn[1:], '0')
	} else {
		return false
	}
}

func TestConsec(t *testing.T) {
	if consec("123", "1234") {
		t.Fail()
	}

	if !consec("123", "124") {
		t.Fail()
	}

	if !consec("1233999", "1234000") {
		t.Fail()
	}
}

func TestIntegration(t *testing.T) {
	rand.Seed(94133)

	fmt.Println("Creating Kinesis")
	kin := kinesis.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		Region: aws.String(os.Getenv("AWS_REGION")),
	})

	stream := os.Getenv("AWS_KINESIS_STREAM")

	redisPool, err := redispool.NewRedisPool(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}
	redisPrefix := os.Getenv("REDIS_PREFIX")

	cpOpt := redischeckpointer.Options{
		SavePeriod:  8 * time.Second,
		RedisPool:   redisPool,
		RedisPrefix: redisPrefix,
	}
	cp, err := redischeckpointer.New(&cpOpt)
	if err != nil {
		panic(err)
	}

	provOpt := redisprovisioner.Options{
		TTL:         8 * time.Second,
		RedisPool:   redisPool,
		RedisPrefix: redisPrefix,
	}
	provOpt2 := provOpt
	prov, err := redisprovisioner.New(&provOpt)
	if err != nil {
		panic(err)
	}

	kinOpt := KinesumerOptions{
		ListStreamsLimit:    25,
		DescribeStreamLimit: 25,
		GetRecordsLimit:     25,
		PollTime:            1,
		MaxShardWorkers:     2,
		Handlers:            DefaultHandlers{},
		DefaultIteratorType: "TRIM_HORIZON",
	}

	fmt.Println("Creating Kinesumer")
	k, err := NewKinesumer(kin, cp, prov, nil, stream, &kinOpt)
	if err != nil {
		panic(err)
	}

	exists, err := k.StreamExists()
	if err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Deleting stream")
		_, err := kin.DeleteStream(&kinesis.DeleteStreamInput{
			StreamName: aws.String(stream),
		})
		if err != nil {
			panic(err)
		}

		for i := 0; i < 60; i++ {
			time.Sleep(time.Second)
			if exists, err := k.StreamExists(); err != nil {
				panic(err)
			} else if !exists {
				goto cont1
			}
		}
		panic("Could not delete stream")
	}
cont1:

	fmt.Println("Creating stream")
	_, err = kin.CreateStream(&kinesis.CreateStreamInput{
		ShardCount: aws.Int64(3),
		StreamName: aws.String(stream),
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		if exists, err := k.StreamExists(); err != nil {
			panic(err)
		} else if exists {
			goto cont2
		}
	}
	panic("Could not create stream")
cont2:

	time.Sleep(time.Second)
	workers, err := k.Begin()
	if err != nil {
		panic(err)
	}

	time.Sleep(8 * time.Second)

	if len(workers) != 2 {
		panic(fmt.Sprintf("Expected 2 workers to be started by k. Workers: %v",
			workers,
		))
	}

	cp2, err := redischeckpointer.New(&cpOpt)
	if err != nil {
		panic(err)
	}

	prov2, err := redisprovisioner.New(&provOpt2)
	if err != nil {
		panic(err)
	}

	kinOpt2 := kinOpt
	kinOpt2.MaxShardWorkers = 3
	k2, err := NewKinesumer(kin, cp2, prov2, nil, stream, &kinOpt2)
	if err != nil {
		panic(err)
	}

	workers2, err := k2.Begin()
	if err != nil {
		panic(err)
	}

	if len(workers2) != 1 {
		panic(fmt.Sprintf("Expected 1 worker to be started by k2. Workers: %v", workers2))
	}

	records := make([]*kinesis.PutRecordsRequestEntry, 100)
	values := make(map[string]bool)
	for i := 0; i < 100; i++ {
		item := uuid.New()
		values[item] = true
		records[i] = &kinesis.PutRecordsRequestEntry{
			Data:         []byte(item),
			PartitionKey: aws.String(item),
		}
	}

	res, err := kin.PutRecords(&kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String(stream),
	})
	if err != nil {
		panic(err)
	}
	if aws.Int64Value(res.FailedRecordCount) != 0 {
		panic(fmt.Sprintf("Failed to put records: %v", res.Records))
	}

	timeout := time.NewTimer(time.Minute)
	for count := 0; count < 100; count++ {
		select {
		case rec := <-k.Records():
			if !values[string(rec.Data())] {
				panic(fmt.Sprintf("Received unexpected record %v", rec))
			} else {
				delete(values, string(rec.Data()))
			}
		case rec := <-k2.Records():
			if !values[string(rec.Data())] {
				panic(fmt.Sprintf("Received unexpected record %v", rec))
			} else {
				delete(values, string(rec.Data()))
			}
		case <-timeout.C:
			panic(fmt.Sprintf("Timed out fetching records. Missing: %v", values))
		}
	}

	if len(values) > 0 {
		panic(fmt.Sprintf("Did not receive all expected records. Missing: %v", values))
	}

	fmt.Println("Basic functionality works")

	stopC := make(chan struct{}, 2)

	smallRecords := records[:10]

	go func() {
		for {
			select {
			case <-stopC:
			default:
				time.Sleep(time.Second)
				res, err := kin.PutRecords(&kinesis.PutRecordsInput{
					Records:    smallRecords,
					StreamName: aws.String(stream),
				})
				if err != nil {
					panic(err)
				}
				if aws.Int64Value(res.FailedRecordCount) != 0 {
					panic(fmt.Sprintf("Failed to put records: %v", res.Records))
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-k.Records():
			case <-k2.Records():
			case <-stopC:
				return
			}
		}
	}()

	shards, err := k.GetShards()
	if err != nil {
		panic(err)
	}
	if len(shards) != 3 {
		panic("Expected 3 shards")
	}

	type pair struct {
		begin, end int
	}
	pairs := []pair{
		{0, 1},
		{0, 2},
		{1, 2},
		{2, 1},
	}
	for _, pair := range pairs {
		if consec(*shards[pair.begin].HashKeyRange.EndingHashKey,
			*shards[pair.end].HashKeyRange.StartingHashKey) {
			_, err := kin.MergeShards(&kinesis.MergeShardsInput{
				ShardToMerge:         shards[pair.begin].ShardID,
				AdjacentShardToMerge: shards[pair.end].ShardID,
				StreamName:           aws.String(stream),
			})
			if err != nil {
				panic(err)
			}
			goto cont3
		}
	}
	panic(func() string {
		s := "Could not find shard to close. Shards: "
		shards, err := k.GetShards()
		if err != nil {
			return err.Error()
		}
		for _, shard := range shards {
			s += shard.GoString()
		}
		return s
	}(),
	)
cont3:

	timeout.Reset(time.Minute)
	select {
	case <-timeout.C:
		panic("Shard worker did not stop after shard closed")
	case <-k.stopped:
		k.stopped <- Unit{}
	case <-k2.stopped:
		k2.stopped <- Unit{}
	}

	stopC <- struct{}{}
	stopC <- struct{}{}

	k.End()
	k2.End()
}
