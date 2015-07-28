package main

import (
	"github.com/codegangsta/cli"
)

var (
	fRedisURL = "redis.url"
)

var flagsRedis = []cli.Flag{
	cli.StringFlag{
		Name:   fRedisURL,
		Usage:  "The Redis URL",
		EnvVar: "REDIS_URL",
	},
}

func getRedisURL(ctx *cli.Context) string {
	return ctx.String("redis.url")
}
