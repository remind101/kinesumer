package redischeckpointer

import (
	"net/url"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	k "github.com/remind101/kinesumer/interface"
)

type TestHandlers struct{}

func (h TestHandlers) Go(f func()) {
	go f()
}

func (h TestHandlers) Err(e k.Error) {
	panic(e)
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		MaxActive:   100,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// redis://x:passwd@host:port
func NewRedisPool(connstr string) (*redis.Pool, error) {

	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	// auth if necessary
	passwd := ""
	if u.User != nil {
		passwd, _ = u.User.Password()
	}

	pool := newPool(u.Host, passwd)

	return pool, nil
}

var (
	prefix      = "pusherman360:testing"
	sequenceKey = prefix + ":sequence"
)

func makeCheckpointer() (*Checkpointer, error) {
	pool, err := NewRedisPool("redis://127.0.0.1:6379")
	if err != nil {
		return nil, err
	}
	r, err := NewRedisCheckpointer(&CheckpointerOptions{
		SavePeriod:  time.Hour,
		RedisPool:   pool,
		RedisPrefix: prefix,
	})
	return r, err
}

func makeCheckpointerWithSamples() *Checkpointer {
	r, _ := makeCheckpointer()
	conn := r.pool.Get()
	defer conn.Close()
	conn.Do("DEL", sequenceKey)
	conn.Do("HSET", sequenceKey, "shard1", "1000")
	conn.Do("HSET", sequenceKey, "shard2", "2000")
	r, _ = makeCheckpointer()
	return r
}

func TestRedisGoodLogin(t *testing.T) {
	r, err := makeCheckpointer()
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
	r := makeCheckpointerWithSamples()
	err := r.Begin(TestHandlers{})
	if err != nil {
		t.Error(err)
	}
	r.End()
}

func TestGetStartSequence(t *testing.T) {
	r := makeCheckpointerWithSamples()
	_ = r.Begin(TestHandlers{})
	r.End()
	shard1 := "shard1"
	seq := r.GetStartSequence(shard1)
	if seq != "1000" {
		t.Error("Expected nonempty sequence number")
	}
}

func TestWriteAll(t *testing.T) {
	r := makeCheckpointerWithSamples()
	r.Begin(TestHandlers{})
	r.heads["shard1"] = "1001"
	r.heads["shard2"] = "2001"
	r.Sync()
	r.End()
	r, _ = makeCheckpointer()
	r.Begin(TestHandlers{})
	r.End()
	if r.heads["shard1"] != "1001" {
		t.Error("Expected sequence number to be written")
	}
}
