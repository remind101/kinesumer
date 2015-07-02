package kinesumer

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/remind101/pkg/logger"
)

type RedisStateSync struct {
	heads    map[string]string
	c        chan *KinesisRecord
	mut      sync.Mutex
	pool     *redis.Pool
	redisKey string
	ticker   <-chan time.Time
	logger   logger.Logger
	wg       sync.WaitGroup
	modified bool
}

type RedisStateSyncOptions struct {
	ShardStateSyncOptions
	RedisURL string
	RedisKey string
}

func NewRedisStateSync(opt *RedisStateSyncOptions) (*RedisStateSync, error) {
	redisPool, err := NewRedisPool(opt.RedisURL)
	if err != nil {
		return nil, err
	}
	return &RedisStateSync{
		heads:    make(map[string]string),
		c:        make(chan *KinesisRecord),
		mut:      sync.Mutex{},
		pool:     redisPool,
		redisKey: opt.RedisKey,
		ticker:   opt.Ticker,
		logger:   opt.Logger,
		modified: true,
	}, nil
}

func (r *RedisStateSync) DoneC() chan *KinesisRecord {
	return r.c
}

func (r *RedisStateSync) Sync() {
	r.logger.Info("Writing sequence numbers")
	r.mut.Lock()
	defer r.mut.Unlock()
	conn := r.pool.Get()
	if len(r.heads) > 0 && r.modified {
		if _, err := conn.Do("HMSET", redis.Args{r.redisKey}.AddFlat(r.heads)...); err != nil {
			r.logger.Error("Failed to sync sequence numbers", "error", err)
		}
		r.modified = false
	} else {
		r.logger.Info("No sequence numbers to write")
	}
}

func (r *RedisStateSync) RunShardSync() {
loop:
	for {
		select {
		case <-r.ticker:
			r.Sync()
		case state, ok := <-r.c:
			if !ok {
				break loop
			}
			r.mut.Lock()
			r.heads[*state.ShardID] = *state.Record.SequenceNumber
			r.modified = true
			r.mut.Unlock()
		}
	}
	r.Sync()
	r.wg.Done()
}

func (r *RedisStateSync) Begin() error {
	conn := r.pool.Get()
	defer conn.Close()
	res, err := conn.Do("HGETALL", r.redisKey)
	r.heads, err = redis.StringMap(res, err)
	if err != nil {
		return err
	}

	r.wg.Add(1)
	go r.RunShardSync()
	return nil
}

func (r *RedisStateSync) End() {
	close(r.c)
	r.wg.Wait()
}

func (r *RedisStateSync) GetStartSequence(shardID *string) *string {
	val, ok := r.heads[*shardID]
	if ok {
		return &val
	} else {
		return nil
	}
}
