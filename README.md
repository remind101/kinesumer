Kinesumer
===
[![Circle CI](https://circleci.com/gh/remind101/kinesumer.svg?style=svg&circle-token=ab11c0337d5aa1aca644e0420b228e86eecdd862)](https://circleci.com/gh/remind101/kinesumer)

Kinesumer is a simple [Go](http://golang.org/) client library for Amazon AWS [Kinesis](http://aws.amazon.com/kinesis/). It aims to be a native Go alternative to Amazon's [KCL](https://github.com/awslabs/amazon-kinesis-client). NOTE: The kinesumer API is nowhere near stable right now so you probably shouldn't use it. However the kinesumer tool should be fine (if it compiles).

Features
---
* Automatically manages one consumer goroutine per shard.
* Handles shard splitting and merging properly.
* Provides a simple channel interface for incoming Kinesis records.
* Provides a tool for managing Kinesis streams:
	* Tailing a stream
	* Splitting and merging shards (Not implemented)

Usage
---
Install
```bash
go get github.com/remind101/kinesumer
```

Example Program
```golang
package main

import (
	"fmt"
	"os"

	"github.com/remind101/kinesumer"
)

func main() {
	k, err := kinesumer.NewDefaultKinesumer(
		"Stream",
	)
	if err != nil {
		panic(err)
	}
	k.Begin()
	defer k.End()
	for i := 0; i < 100; i++ {
		rec := <-k.Records()
		fmt.Println(string(rec.Data()))
	}
}
```
