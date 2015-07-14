Kinesumer
===

Kinesumer is a simple [Go](http://golang.org/) client library for Amazon AWS [Kinesis](http://aws.amazon.com/kinesis/). It aims to be a native Go alternative to Amazon's [KCL](https://github.com/awslabs/amazon-kinesis-client).

Features
---
* Automatically manages one consumer goroutine per shard.
* Handles shard splitting and merging properly.
* Provides a simple channel interface for incoming Kinesis records.

Usage
---
Install
```bash
go get github.com/remind101/kinesumer
```
TODO: example consumer
