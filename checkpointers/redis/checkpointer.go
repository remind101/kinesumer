package redischeckpointer

import (
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	k "github.com/remind101/kinesumer/interface"
)

type Checkpointer struct {
	heads       map[string]string
	c           chan k.Record
	mut         sync.Mutex
	pool        *redis.Pool
	redisPrefix string
	savePeriod  time.Duration
	wg          sync.WaitGroup
	modified    bool
	handlers    k.Handlers
}

type CheckpointerOptions struct {
	SavePeriod  time.Duration
	RedisPool   *redis.Pool
	RedisPrefix string
}

func NewRedisCheckpointer(opt *CheckpointerOptions) (*Checkpointer, error) {
	return &Checkpointer{
		heads:       make(map[string]string),
		c:           make(chan k.Record),
		mut:         sync.Mutex{},
		pool:        opt.RedisPool,
		redisPrefix: opt.RedisPrefix,
		savePeriod:  opt.SavePeriod,
		modified:    true,
	}, nil
}

func (r *Checkpointer) DoneC() chan<- k.Record {
	return r.c
}

func (r *Checkpointer) Sync() {
	r.mut.Lock()
	defer r.mut.Unlock()
	if len(r.heads) > 0 && r.modified {
		conn := r.pool.Get()
		defer conn.Close()
		if _, err := conn.Do("HMSET", redis.Args{r.redisPrefix + ":sequence"}.AddFlat(r.heads)...); err != nil {
			// TODO: report err
		}
		r.modified = false
	}
}

func (r *Checkpointer) RunCheckpointer() {
	saveTicker := time.NewTicker(r.savePeriod).C
loop:
	for {
		select {
		case <-saveTicker:
			r.Sync()
		case state, ok := <-r.c:
			if !ok {
				break loop
			}
			r.mut.Lock()
			r.heads[state.ShardID()] = state.SequenceNumber()
			r.modified = true
			r.mut.Unlock()
		}
	}
	r.Sync()
	r.wg.Done()
}

func (r *Checkpointer) Begin(handlers k.Handlers) error {
	r.handlers = handlers

	conn := r.pool.Get()
	defer conn.Close()
	res, err := conn.Do("HGETALL", r.redisPrefix+":sequence")
	r.heads, err = redis.StringMap(res, err)
	if err != nil {
		return err
	}

	r.wg.Add(1)
	r.handlers.Go(func() {
		r.RunCheckpointer()
	})
	return nil
}

func (r *Checkpointer) End() {
	close(r.c)
	r.wg.Wait()
}

func (r *Checkpointer) GetStartSequence(shardID *string) *string {
	val, ok := r.heads[*shardID]
	if ok {
		return &val
	} else {
		return nil
	}
}
