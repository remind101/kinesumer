// +build integration

package kinesumer

import (
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

func TestEverything(t *testing.T) {
	resetTestHandlers()

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

	if !exists {
		t.Logf("Stream does not exist. Creating.")
		_, err := kin.CreateStream(&kinesis.CreateStreamInput{
			ShardCount: aws.Int64(3),
			StreamName: aws.String(stream),
		})
		if err != nil {
			panic(err)
		}

		for i := 0; i < 60; i++ {
			time.Sleep(100 * time.Millisecond)
			if exists, err := k.StreamExists(); err != nil {
				panic(err)
			} else if exists {
				goto cont
			}
		}

		panic("Could not create stream")
	} else {
		t.Logf("Stream already exists")
	}

cont:
	workers, err := k.Begin()
	if err != nil {
		panic(err)
	}

	for _, proc := range toRun {
		go proc()
	}

	if len(workers) != 2 {
		panic("Expected 2 workers to be started by k")
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

	t.FailNow()
}
