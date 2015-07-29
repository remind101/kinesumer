package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/remind101/kinesumer"
	"github.com/remind101/kinesumer/checkpointers/redis"
	"github.com/remind101/kinesumer/redispool"
)

var cmdTail = cli.Command{
	Name:    "tail",
	Aliases: []string{"t"},
	Usage:   "Pipes a Kinesis stream to standard out",
	Action:  runTail,
	Flags: append(
		[]cli.Flag{
			cli.StringFlag{
				Name:  "stream, s",
				Usage: "The Kinesis stream to tail",
			},
		}, flagsAWSRedis...,
	),
}

func runTail(ctx *cli.Context) {
	k, err := kinesumer.NewDefaultKinesumer(
		ctx.String(fAWSAccess),
		ctx.String(fAWSSecret),
		ctx.String(fAWSRegion),
		ctx.String("stream"),
	)
	if err != nil {
		panic(err)
	}

	if redisURL := ctx.String(fRedisURL); len(redisURL) > 0 {
		pool, err := redispool.NewRedisPool(redisURL)
		if err != nil {
			panic(err)
		}

		cp, err := redischeckpointer.New(&redischeckpointer.Options{
			ReadOnly:    true,
			RedisPool:   pool,
			RedisPrefix: ctx.String(fRedisPrefix),
		})
		if err != nil {
			panic(err)
		}

		k.Checkpointer = cp
	}

	_, err = k.Begin()
	if err != nil {
		panic(err)
	}
	defer k.End()
	for {
		rec := <-k.Records()
		fmt.Println(string(rec.Data()))
	}
}
