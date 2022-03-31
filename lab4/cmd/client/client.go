package main

import (
	"context"
	"flag"
	"log"
	"strconv"
	"time"

	"cs426.yale.edu/lab4/kv"
)

// Simple CLI for interacting with a KV cluster. You can get, set, or delete values using the Kv API.
// You must pass in a shardmap as a JSON file with --shardmap=path/to/shardmap.json
//
// Examples:
//   - go run cmd/client/client.go --shardmap=shardmaps/test-1.json set abc 123 5000  # sets "abc" to "123" with TTL of 5s (5000ms)
//   - go run cmd/client/client.go --shardmap=shardmaps/test-1.json get abc           # retrieves value for key "abc" (should be "123")
//   - go run cmd/client/client.go --shardmap=shardmaps/test-1.json delete abc        # removes value at "abc"

var (
	shardMapFile = flag.String("shardmap", "", "Path to a JSON file which describes the shard map")
)

func usage() {
	log.Fatal("Usage: client.go [get|set|delete] key [value] [ttl]")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		usage()
	}

	fileSm, err := kv.WatchShardMapFile(*shardMapFile)
	if err != nil {
		log.Fatal(err)
	}

	clientPool := kv.MakeClientPool(&fileSm.ShardMap)

	client := kv.MakeKv(&fileSm.ShardMap, &clientPool)

	subcommand := args[0]
	key := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	switch subcommand {
	case "get":
		value, isSet, err := client.Get(ctx, key)
		if err != nil {
			log.Printf("error getting value for key: %q", err)
		} else if !isSet {
			log.Printf("no value set for key: %s", key)
		} else {
			println(value)
		}
	case "set":
		if len(args) < 4 {
			usage()
		}
		value := args[2]
		ttlMs, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			log.Fatalf("expected int value for ttlMs: %q", err)
		}
		err = client.Set(ctx, key, value, time.Duration(ttlMs)*time.Millisecond)
		if err != nil {
			log.Printf("error setting value for key: %q", err)
		}
	case "delete":
		err := client.Delete(ctx, key)
		if err != nil {
			log.Printf("error deleting value for key: %q", err)
		}
	default:
		usage()
	}
	cancel()
}
