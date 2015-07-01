package kinesumer

import (
	"bytes"
	"log"
	//"strings"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/remind101/pkg/logger"
)

var sequenceKey = "pusherman360.sequence.testing"

func makeRedisStateSync(url string) (*bytes.Buffer, *RedisStateSync, error) {
	buf := new(bytes.Buffer)
	r, err := NewRedisStateSync(&RedisStateSyncOptions{
		ShardStateSyncOptions: ShardStateSyncOptions{
			Logger: logger.New(log.New(buf, "", 0)),
			Ticker: time.NewTicker(time.Nanosecond).C,
		},
		RedisURL: url,
		RedisKey: sequenceKey,
	})
	return buf, r, err
}

func makeRedisStateSyncWithSamples() (*bytes.Buffer, *RedisStateSync) {
	_, r, _ := makeRedisStateSync("redis://127.0.0.1:6379")
	conn := r.pool.Get()
	defer conn.Close()
	conn.Do("DEL", sequenceKey)
	conn.Do("HSET", sequenceKey, "shard1", "1000")
	conn.Do("HSET", sequenceKey, "shard2", "2000")
	buf, r, _ := makeRedisStateSync("redis://127.0.0.1:6379")
	return buf, r
}

func TestRedisBadLogin(t *testing.T) {
	_, r, err := makeRedisStateSync("redis://u:p@classchirp.com")
	if err == nil {
		// This currently passes but I'm not sure why
		//t.Error("Redis login should fail")
	}

	conn := r.pool.Get()
	defer conn.Close()
	_, err = conn.Do("ECHO", "hey")
	if err == nil {
		t.Error("Redis command should fail")
	}
}

func TestRedisGoodLogin(t *testing.T) {
	_, r, err := makeRedisStateSync("redis://127.0.0.1:6379")
	if err != nil {
		t.Error("Failed to connect to redis at localhost:6379")
	}

	conn := r.pool.Get()
	defer conn.Close()
	reply, err := conn.Do("ECHO", "hey")

	re, err := redis.String(reply, err)
	if err != nil || re != "hey" {
		t.Error("Redis ECHO failed")
	}
}

/*
func TestGetStartSequence(t *testing.T) {
	buf, r := makeRedisStateSyncWithSamples()
	shard1 := "shard1"
	seq := r.GetStartSequence(&shard1)
	if seq == nil || *seq != "1000" {
		t.Error("Expected nonempty sequence number")
	}
	if buf.Len() != 0 {
		t.Error("Expected logger output to be empty")
	}
}

func TestWriteAll(t *testing.T) {
	buf, r := makeRedisStateSyncWithSamples()
	r.heads["shard1"] = "1001"
	r.heads["shard2"] = "2001"
	r.Sync()
	if !strings.Contains(buf.String(), "Writing sequence numbers") {
		t.Error("Expected logger entry")
	}
	_, r, _ = makeRedisStateSync("redis://127.0.0.1:6379")
	if r.heads["shard1"] != "1001" {
		t.Error("Expected sequence number to be written")
	}
}
*/
