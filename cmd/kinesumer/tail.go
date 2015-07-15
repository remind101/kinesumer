package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/remind101/kinesumer"
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
		}, flagsAws...,
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
	k.Begin()
	defer k.End()
	for {
		rec := <-k.Records()
		if rec.Err != nil {
			fmt.Fprintf(os.Stdout, "%v\n", rec.Err.Error())
			if rec.ShardID != nil {
				fmt.Fprintf(os.Stdout, "at shard %v\n", *rec.ShardID)
			}
		} else {
			fmt.Println(string(rec.Data))
		}
	}
}
