package redisprovisioner

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pborman/uuid" // Exported from code.google.com/p/go-uuid
)

type Provisioner struct {
	acquired    map[string]bool
	alivePeriod time.Duration
	pool        *redis.Pool
	redisPrefix string
	lock        string
}

func New() (*Provisioner, error) {
	return &Provisioner{
		acquired:    make(map[string]bool),
		alivePeriod: 0,
		pool:        nil,
		redisPrefix: "",
		lock:        uuid.New(),
	}, nil
}

func (p *Provisioner) Begin() {
}

func (p *Provisioner) End() {
}

func (p *Provisioner) TryAcquire(shardID *string) error {
	conn := p.pool.Get()
	defer conn.Close()
	if p.acquired[*shardID] {
		return errors.New("Lock already acquired by this process")
	}
	res, err := conn.Do("SET", p.redisPrefix+".lock."+*shardID, p.lock, "PX", p.alivePeriod/time.Millisecond, "NX")
	if err != nil {
		return err
	}
	if res != "OK" {
		return errors.New("Failed to acquire lock")
	}
	p.acquired[*shardID] = true
	return nil
}

func (p *Provisioner) Refresh() {
	conn := p.pool.Get()
	defer conn.Close()
	for shardID := range p.acquired {
		res, err := conn.Do("PEXPIRE", p.redisPrefix+".lock."+shardID, p.alivePeriod/time.Millisecond, "NX")
		if err != nil || res != "OK" {
			if err != nil {
				panic(err) // TODO: error handling
			}
			err = p.TryAcquire(&shardID)
			if err != nil {
				panic(err) // TODO: error handling
			}
		}
	}
}

func (p *Provisioner) Release(shardID *string) error {
	conn := p.pool.Get()
	defer conn.Close()
	delete(p.acquired, *shardID)
	key := p.redisPrefix + ".lock." + *shardID
	res, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}
	if res != p.lock {
		return errors.New("Bad lock")
	}
	_, err = conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}
