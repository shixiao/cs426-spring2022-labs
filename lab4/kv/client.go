package kv

import (
	"context"
	"time"
)

type Kv struct {
	shardMap   *ShardMap
	clientPool ClientPool

	// Add any client-side state you want here
}

func MakeKv(shardMap *ShardMap, clientPool ClientPool) *Kv {
	kv := &Kv{
		shardMap:   shardMap,
		clientPool: clientPool,
	}
	// Add any initialization logic
	return kv
}

func (kv *Kv) Get(ctx context.Context, key string) (string, bool, error) {
	panic("TODO: Part B")
}

func (kv *Kv) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	panic("TODO: Part B")
}

func (kv *Kv) Delete(ctx context.Context, key string) error {
	panic("TODO: Part B")
}
