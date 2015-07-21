package kinesumer

import (
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	prefix      = "pusherman360:testing"
	sequenceKey = prefix + ":sequence"
)

func makeRedisStateSync() (*RedisStateSync, error) {
	r, err := NewRedisStateSync(&RedisStateSyncOptions{
		SavePeriod:  time.Hour,
		AlivePeriod: time.Hour,
		RedisURL:    "redis://127.0.0.1:6379",
		RedisPrefix: prefix,
	})
	return r, err
}

func makeRedisStateSyncWithSamples() *RedisStateSync {
	r, _ := makeRedisStateSync()
	conn := r.pool.Get()
	defer conn.Close()
	conn.Do("DEL", sequenceKey)
	conn.Do("HSET", sequenceKey, "shard1", "1000")
	conn.Do("HSET", sequenceKey, "shard2", "2000")
	r, _ = makeRedisStateSync()
	return r
}

func TestRedisGoodLogin(t *testing.T) {
	r, err := makeRedisStateSync()
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

func TestRedisBeginEnd(t *testing.T) {
	r := makeRedisStateSyncWithSamples()
	c := make(chan *KinesisRecord)
	err := r.Begin(c)
	if err != nil {
		t.Error(err)
	}
	r.End()
}

func TestGetStartSequence(t *testing.T) {
	r := makeRedisStateSyncWithSamples()
	c := make(chan *KinesisRecord)
	_ = r.Begin(c)
	r.End()
	shard1 := "shard1"
	seq := r.GetStartSequence(&shard1)
	if seq == nil || *seq != "1000" {
		t.Error("Expected nonempty sequence number")
	}
}

func TestWriteAll(t *testing.T) {
	r := makeRedisStateSyncWithSamples()
	c := make(chan *KinesisRecord)
	r.Begin(c)
	r.heads["shard1"] = "1001"
	r.heads["shard2"] = "2001"
	r.Sync()
	r.End()
	r, _ = makeRedisStateSync()
	r.Begin(c)
	r.End()
	if r.heads["shard1"] != "1001" {
		t.Error("Expected sequence number to be written")
	}
}
