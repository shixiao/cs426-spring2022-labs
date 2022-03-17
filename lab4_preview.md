# [Preview] Lab 4: Sharded Key-Value Cache

In this lab you will build a sharded and replicated key-value cache that uses
logical sharding with hash sharding of keys. This will be split roughly three
parts:
 - Implementing a simple gRPC server which handles `get(key)` requests and `set(key, value, TTL)` requests. This server will store data in-memory in a thread-safe way.
 - Implement a client library, which routes requests for keys to the appropriate backing gRPC server using a provided `ShardMap` configuration.
 - Implement *shard migrations* in the gRPC server by copying data from live replicas to allow for dynamic shard movements.

## Background

We have provided a `ShardMap` interfaces which mimics a cluster configuration. It provides
the list of nodes and their connection information (i.e. their IP/port) as well as the mapping
of shards to host. For example:

```
{
    "nodes": {
        "n1": "127.0.0.1:9001",
        "n2": "127.0.0.1:9002",
        "n3": "127.0.0.1:9003"
    },
    "shards": 5,
    "shardMapping": {
        "n1": [1, 2, 3, 5],
        "n2": [1, 2, 4],
        "n3": [2, 3, 4]
    }
}
```

This configures a cluster with 3 nodes and 5 shards. Shard 1 maps to *both* node 1 and node 2 (i.e. it has a replication factor of 2) -- so all writes to shard 1 must be sent to both
N1 and N2. Reads to shard 1 may be sent to either node 1 or node 2 -- you will implement
primitive round-robin based load-balancing in **Part B**. Note that in this lab
we will not deal with partial failures or divergent replicas: `KVServer` operates like
a cache so it is ok if some data is missing due to partial failures.

This sharding scheme is *dynamic* -- shards may be moved between nodes.
The provided `ShardMap` implemention will wrap this configuration structure for you and
provide notifications (via a `chan`) when the shard map changes. Your `KVServer`s will
implement a shard migration protocol in **Part C** to handle this without losing data.

## Part A. Server implementation

We have provided a gRPC interface called `KVServer` which you will implement. It consists of
only three methods:
 1. `get(key)` which returns the value stored at key, or `nil` in the case that it does not exist.
 2. `set(key, value, ttl)` which stores the `value` in the in-memory store under the `key`. It should be deleted after `ttl` seconds have past.
 3. `delete(key)` which removes the value stored at key if one exists.
Each node in the cluster will run one instance of `KVServer`, and the client will
route requests appropriately based on the `ShardMap` (see **Part B**).

You may store data in any way you like in-memory as long as it is thread safe and
can appropriately implement the API methods. A few tips:
 - You will need to map the given `key` to a shard in the server, using the provided `ShardMap` instance on the server struct.
 - Store data for separate shards separately to make **Part C** easier.

You may make the data structure thread-safe in any way, but the server must be able
to handle requests with some level of concurrency -- a single global Mutex is not sufficient.
Some suggestions:
 - Use lock striping (from Lab 0) on the shard or key level.
 - Use reader/writer locks to allow reads to happen concurrently if there are no writers.
You may assume that the workload is relatively read heavy
(say at least 2x as many reads as writes) when making rough design decisions, though there
may be some tests which have more writes than reads.

## Part B. Client implementation
We have provided a skeleton of the client library in `kv_client.go`, which outlines the
same 3 methods as the gRPC interface: `get`, `set`, and `delete`. You will implement these
methods by:
 - Hashing the key to get the appropriate shard
 - Finding the appropriate servers for the shard to route to using the `ShardMap`
     - For `get` requests you will pick a single server from the potential list using round-robin
     - For `set`/`delete` requests you will send the request to all servers that host the shard
 - Constructing gRPC requests and sending to those servers
 - Handling errors
     - For `get` requests, you will retry a failed request onto another server that hosts the shard (if any are availanle)
     - For `set`/`delete` requests you will consider the request a failure if any server fails

## Part C. Shard migrations

Up to this point you should have a working `KV` implementation as long as the shard map does
not change. In this step you will implement parts of the dynamic shard movement protocol, including:
 - The server will reject requests to shards it does not host, to prevent misrouted requests if the client has a stale `ShardMap` or races with a shard movement
 - The server will clean up data if a shard is removed
 - The server will *copy* data from other replicas (if any are available)
     - When a new shard is added to a server, it will search for other servers that also host that shard. If any are available it will issue an RPC to fetch all the data for that shard and insert it into its local storage, effectively copying the state.
