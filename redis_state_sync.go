package kinesumer

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/remind101/pkg/logger"
)

type RedisStateSync struct {
	heads    map[string]*string
	c        *chan *ShardState
	mut      sync.Mutex
	pool     *redis.Pool
	redisKey string
	ticker   *<-chan time.Time
	logger   logger.Logger
	wg       sync.WaitGroup
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
	c := make(chan *ShardState)
	return &RedisStateSync{
		heads:    make(map[string]*string),
		c:        &c,
		mut:      sync.Mutex{},
		pool:     redisPool,
		redisKey: opt.RedisKey,
		ticker:   opt.Ticker,
		logger:   opt.Logger,
	}, nil
}

func (r *RedisStateSync) doneC() *chan *ShardState {
	return r.c
}

func (r *RedisStateSync) writeOne(sequence, shardID *string) {
	conn := r.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", r.redisKey, *shardID, *sequence)
	if err != nil {
		r.logger.Error("Failed to write sequence number", "shard", *shardID,
			"sequence", *sequence)
	}
}

func (r *RedisStateSync) writeAll() {
	r.logger.Info("Writing sequence numbers")
	r.mut.Lock()
	defer r.mut.Unlock()
	for id, seq := range r.heads {
		r.writeOne(&id, seq)
	}
}

func (r *RedisStateSync) begin() {
	r.wg.Add(1)
	go func() {
	loop:
		for {
			select {
			case <-*r.ticker:
				r.writeAll()
			case state, ok := <-*r.c:
				if !ok {
					break loop
				}
				r.mut.Lock()
				r.heads[*state.ShardID] = state.Record.SequenceNumber
				r.mut.Unlock()
			}
		}
		r.writeAll()
		r.wg.Done()
	}()
}

func (r *RedisStateSync) end() {
	close(*r.c)
	r.wg.Wait()
}

func (r *RedisStateSync) getStartSequence(shardID *string) *string {
	conn := r.pool.Get()
	defer conn.Close()
	resp, err := conn.Do("HGET", r.redisKey, *shardID)

	tmp, err := redis.String(resp, err)
	if err != nil {
		r.logger.Warn("Error when fetching starting sequence number", "shard", *shardID, "error", err)
		return nil
	}
	return &tmp
}
