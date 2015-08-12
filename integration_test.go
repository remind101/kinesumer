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
	"github.com/remind101/kinesumer/checkpointers/redis"
	"github.com/remind101/kinesumer/provisioners/redis"
	"github.com/remind101/kinesumer/redispool"
)

type NSync struct {
	In  chan struct{}
	Out chan struct{}
}

func NewNSync(n int) NSync {
	sync := NSync{
		In:  make(chan struct{}),
		Out: make(chan struct{}),
	}
	go func() {
		for {
			for i := 0; i < n; i++ {
				_, ok := <-sync.In
				if !ok {
					if i > 0 {
						panic("NSync stream closed")
					}
					close(sync.Out)
					return
				}
			}
			for i := 0; i < n; i++ {
				sync.Out <- struct{}{}
			}
		}
	}()
	return sync
}

func (n NSync) Sync() {
	n.In <- struct{}{}
	<-n.Out
}

func TestIntegration(t *testing.T) {
	rand.Seed(94133)

	t.Logf("Creating Kinesis")
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
		SavePeriod:  time.Second,
		RedisPool:   redisPool,
		RedisPrefix: redisPrefix,
	}
	cp, err := redischeckpointer.New(&cpOpt)
	if err != nil {
		panic(err)
	}

	provOpt := redisprovisioner.Options{
		TTL:         time.Second,
		RedisPool:   redisPool,
		RedisPrefix: redisPrefix,
	}
	prov, err := redisprovisioner.New(&provOpt)
	if err != nil {
		panic(err)
	}

	kinOpt := KinesumerOptions{
		ListStreamsLimit:    25,
		DescribeStreamLimit: 25,
		GetRecordsLimit:     1,
		PollTime:            1,
		MaxShardWorkers:     2,
		Handlers:            DefaultHandlers{},
	}

	t.Logf("Creating Kinesumer")
	k, err := NewKinesumer(kin, cp, prov, nil, stream, &kinOpt)
	if err != nil {
		panic(err)
	}

	exists, err := k.StreamExists()
	if err != nil {
		panic(err)
	}

	if exists {
		t.Logf("Deleting stream")
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

	t.Logf("Creating stream")
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

	workers, err := k.Begin()
	if err != nil {
		panic(err)
	}

	if len(workers) != 2 {
		panic(fmt.Sprintf("Expected 2 workers to be started by k. Workers: %v",
			workers,
		))
	}

	cp2, err := redischeckpointer.New(&cpOpt)
	if err != nil {
		panic(err)
	}

	prov2, err := redisprovisioner.New(&provOpt)
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
		panic("Expected 1 worker to be started by k2")
	}

	records := make([]*PutRecordsRequestEntry, 100)
	for i := 0; i < 100; i++ {
		records[i] = &kinesis.PutRecordsRequestEntry{
			Data:         []byte(""),
			PartitionKey: aws.String("aeou"),
		}
	}

	kin.PutRecords(&kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String(stream),
	})

	t.FailNow()
}
