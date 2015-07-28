package main

import (
	"fmt"

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
	err = k.Begin()
	if err != nil {
		panic(err)
	}
	defer k.End()
	for {
		rec := <-k.Records()
		fmt.Println(string(rec.Data()))
	}
}
