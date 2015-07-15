package main

import (
	"github.com/codegangsta/cli"
)

var (
	fAWSRegion = "aws.region"
	fAWSAccess = "aws.accesskey"
	fAWSSecret = "aws.secretkey"
)

var flagsAws = []cli.Flag{
	cli.StringFlag{
		Name:   fAWSRegion,
		Value:  "us-east-1",
		Usage:  "The AWS Kinesis region",
		EnvVar: "AWS_REGION",
	},
	cli.StringFlag{
		Name:   fAWSAccess,
		Value:  "AAAAAAAAAAAAAAAAAAAA",
		Usage:  "The AWS access key",
		EnvVar: "AWS_ACCESS_KEY_ID",
	},
	cli.StringFlag{
		Name:   fAWSSecret,
		Value:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Usage:  "The AWS secret key",
		EnvVar: "AWS_SECRET_ACCESS_KEY",
	},
}
