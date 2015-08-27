package main

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/remind101/kinesumer"
)

var cmdShards = cli.Command{
	Name:    "shards",
	Aliases: []string{"sh"},
	Usage:   "Gets the shards of a stream",
	Action:  runShards,
	Flags:   flagsStream,
}

type nums []*big.Int

func (n nums) Len() int {
	return len(n)
}

func (n nums) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nums) Less(i, j int) bool {
	return n[i].Cmp(n[j]) < 0
}

func (n *nums) UniqSort() {
	sort.Sort(n)

	tmp := make(nums, 0)
	for _, key := range *n {
		if len(tmp) == 0 || tmp[len(tmp)-1].Cmp(key) != 0 {
			tmp = append(tmp, key)
		}
	}

	*n = tmp
}

func (n nums) Clone() nums {
	p := make(nums, len(n))
	for i := 0; i < len(n); i++ {
		p[i] = &big.Int{}
		p[i].Set(n[i])
	}
	return p
}

func runShards(ctx *cli.Context) {
	stream := getStream(ctx)
	k, err := kinesumer.NewDefault(
		stream,
	)
	if err != nil {
		panic(err)
	}

	shards, err := k.GetShards()
	if err != nil {
		panic(err)
	}
	if len(shards) == 0 {
		fmt.Printf("No shards found on stream %s\n", stream)
	}

	keys := make(nums, 0)

	for _, shard := range shards {
		begin := &big.Int{}
		begin.SetString(*shard.HashKeyRange.StartingHashKey, 10)
		keys = append(keys, begin)

		end := &big.Int{}
		end.SetString(*shard.HashKeyRange.EndingHashKey, 10)
		end.Add(end, big.NewInt(1))
		keys = append(keys, end)
	}

	keys.UniqSort()

	// Only support < 100 shards for now
	maxShardIdLen := 0
	for _, shard := range shards {
		if maxShardIdLen < len(*shard.ShardId) {
			maxShardIdLen = len(*shard.ShardId)
		}
	}

	fmt.Printf("SHARD ID%s  ", strings.Repeat(" ", maxShardIdLen-len("SHARD ID")))
	for i := 0; i < len(keys); i++ {
		if i < 10 {
			fmt.Printf("%d  ", i)
		} else {
			fmt.Printf("%d ", i)
		}
	}
	fmt.Println()

	for _, shard := range shards {
		fmt.Printf("%s%s  ", *shard.ShardId, strings.Repeat(" ", maxShardIdLen-len(*shard.ShardId)))
		for _, key := range keys {
			begin := &big.Int{}
			begin.SetString(*shard.HashKeyRange.StartingHashKey, 10)
			if key.Cmp(begin) < 0 {
				fmt.Printf("   ")
				continue
			}
			end := &big.Int{}
			end.SetString(*shard.HashKeyRange.EndingHashKey, 10)
			end.Add(end, big.NewInt(1))
			if key.Cmp(end) < 0 {
				fmt.Printf("o--")
			} else {
				fmt.Printf("o\n")
				break
			}
		}
	}
	fmt.Printf("\nHash keys:\n")
	for i, key := range keys {
		fmt.Printf("%d: %s\n", i, key.String())
	}
}
