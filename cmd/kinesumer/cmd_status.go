package main

import (
	"github.com/codegangsta/cli"
	"github.com/remind101/kinesumer"
	"github.com/remind101/kinesumer/checkpointers/redis"
	"github.com/remind101/kinesumer/redispool"
)

var cmdStatus = cli.Command{
	Name:    "status",
	Aliases: []string{"s"},
	Usage:   "Gets the status of a kinesis stream",
	Action:  runStatus,
	Flags: append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  "stream, s",
				Usage: "The Kinesis stream to tail",
			},
		},
		append(flagsAws, flagsRedis...)...,
	),
}

func runStatus(ctx *cli.Context) {
	k, err := kinesumer.NewDefaultKinesumer(
		ctx.String(fAWSAccess),
		ctx.String(fAWSSecret),
		ctx.String(fAWSRegion),
		ctx.String("stream"),
	)
	if err != nil {
		panic(err)
	}

	pool, err := redispool.NewRedisPool(ctx.String(fRedisURL))
	if err != nil {
		panic(err)
	}

	prefix := ctx.String(fRedisPrefix)

	k.Checkpointer, err = redischeckpointer.NewRedisCheckpointer(&redischeckpointer.CheckpointerOptions{
		ReadOnly:    true,
		RedisPool:   pool,
		RedisPrefix: prefix,
	})
	if err != nil {
		panic(err)
	}

	// TODO: implement
}
