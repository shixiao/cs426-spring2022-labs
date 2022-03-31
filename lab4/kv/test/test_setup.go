package kvtest

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"cs426.yale.edu/lab4/kv"
	"cs426.yale.edu/lab4/kv/proto"
)

type TestSetup struct {
	shardMap   *kv.ShardMap
	nodes      map[string]*kv.KvServerImpl
	clientPool TestClientPool
	kv         *kv.Kv
	ctx        context.Context
}

func MakeTestSetup(shardMap kv.ShardMapState) *TestSetup {
	setup := TestSetup{
		shardMap: &kv.ShardMap{},
		ctx:      context.Background(),
		nodes:    make(map[string]*kv.KvServerImpl),
	}
	setup.shardMap.Update(&shardMap)
	for name := range setup.shardMap.Nodes() {
		setup.nodes[name] = kv.MakeKvServer(
			name,
			setup.shardMap,
			&setup.clientPool,
		)
	}
	setup.clientPool.Setup(setup.nodes)
	setup.kv = kv.MakeKv(setup.shardMap, &setup.clientPool)
	return &setup
}
func MakeTestSetupWithoutServers(shardMap kv.ShardMapState) *TestSetup {
	// Remove nodes so we never have a chance of sending data
	// to the KvServerImpl attached as a safety measure for client_test.go
	//
	// Server tests happen in server_test.go, so we can test
	// client implementations separately.
	//
	// Combined tests happen in integration_tests.go (server and client together),
	// along with stress tests.
	setup := TestSetup{
		shardMap: &kv.ShardMap{},
		ctx:      context.Background(),
		nodes:    make(map[string]*kv.KvServerImpl),
	}
	setup.shardMap.Update(&shardMap)
	for name := range setup.shardMap.Nodes() {
		setup.nodes[name] = nil
	}
	setup.clientPool.Setup(setup.nodes)
	setup.kv = kv.MakeKv(setup.shardMap, &setup.clientPool)
	return &setup
}

func (ts *TestSetup) NodeGet(nodeName string, key string) (string, bool, error) {
	response, err := ts.nodes[nodeName].Get(context.Background(), &proto.GetRequest{Key: key})
	if err != nil {
		return "", false, err
	}
	return response.Value, response.WasFound, nil
}
func (ts *TestSetup) NodeSet(nodeName string, key string, value string, ttl time.Duration) error {
	_, err := ts.nodes[nodeName].Set(
		context.Background(),
		&proto.SetRequest{Key: key, Value: value, TtlMs: ttl.Milliseconds()},
	)
	return err
}
func (ts *TestSetup) NodeDelete(nodeName string, key string) error {
	_, err := ts.nodes[nodeName].Delete(context.Background(), &proto.DeleteRequest{Key: key})
	return err
}

func (ts *TestSetup) Get(key string) (string, bool, error) {
	return ts.kv.Get(ts.ctx, key)
}
func (ts *TestSetup) Set(key string, value string, ttl time.Duration) error {
	return ts.kv.Set(ts.ctx, key, value, ttl)
}
func (ts *TestSetup) Delete(key string) error {
	return ts.kv.Delete(ts.ctx, key)
}

func (ts *TestSetup) UpdateShardMapping(shardsToNodes map[int][]string) {
	state := kv.ShardMapState{
		Nodes:         ts.shardMap.Nodes(),
		NumShards:     ts.shardMap.NumShards(),
		ShardsToNodes: shardsToNodes,
	}
	ts.shardMap.Update(&state)

	// We update again, which effectively waits for the first update to have been processed,
	// because we wait for updates to be processed stricly in order.
	//
	// TODO: decide if we want this or a different sync API
	ts.shardMap.Update(&state)
}

func (ts *TestSetup) NumShards() int {
	return ts.shardMap.NumShards()
}

func nodesExcept(nodes []string, nodeToRemove string) []string {
	newNodes := make([]string, 0)
	for _, node := range nodes {
		if node != nodeToRemove {
			newNodes = append(newNodes, node)
		}
	}
	return newNodes
}

func (ts *TestSetup) DrainNode(nodeToDrain string) {
	shardsToNodes := make(map[int][]string, ts.shardMap.NumShards())
	for shard := 1; shard <= ts.shardMap.NumShards(); shard++ {
		shardsToNodes[shard] = nodesExcept(ts.shardMap.NodesForShard(shard), nodeToDrain)
	}
	ts.UpdateShardMapping(shardsToNodes)
}

func (ts *TestSetup) MoveShard(shardToMove int, srcNode string, dstNode string) {
	if srcNode == dstNode {
		return
	}

	// copy shard to dstNode
	shardsToNodes := make(map[int][]string, ts.shardMap.NumShards())
	for shard := 1; shard <= ts.shardMap.NumShards(); shard++ {
		nodes := ts.shardMap.NodesForShard(shard)
		if shard == shardToMove {
			nodes = append(nodes, dstNode)
		}
		shardsToNodes[shard] = nodes
	}
	ts.UpdateShardMapping(shardsToNodes)

	// remove from srcNode
	shardsToNodes = make(map[int][]string, ts.shardMap.NumShards())
	for shard := 1; shard <= ts.shardMap.NumShards(); shard++ {
		nodes := ts.shardMap.NodesForShard(shard)
		if shard == shardToMove {
			shardsToNodes[shard] = nodesExcept(nodes, srcNode)
		} else {
			shardsToNodes[shard] = nodes
		}
	}
	ts.UpdateShardMapping(shardsToNodes)
}

func (ts *TestSetup) MoveRandomShard() {
	shardToMove := rand.Int()%ts.NumShards() + 1
	nodes := ts.shardMap.NodesForShard(shardToMove)
	nodesMap := make(map[string]struct{}, 0)
	for _, node := range nodes {
		nodesMap[node] = struct{}{}
	}
	srcNode := nodes[rand.Int()%len(nodes)]
	dstCandidate := make([]string, 0)
	for node := range ts.nodes {
		_, present := nodesMap[node]
		if !present {
			dstCandidate = append(dstCandidate, node)
		}
	}
	if len(dstCandidate) == 0 {
		// if the shard is already replicated on all replicas, simply drop from a random node
		ts.DropShardFromNode(shardToMove, srcNode)
	} else {
		dstNode := dstCandidate[rand.Int()%len(dstCandidate)]
		ts.MoveShard(shardToMove, srcNode, dstNode)
	}
}

func (ts *TestSetup) DropShardFromNode(shardToDrop int, srcNode string) {
	shardsToNodes := make(map[int][]string, ts.shardMap.NumShards())
	for shard := 1; shard <= ts.shardMap.NumShards(); shard++ {
		nodes := ts.shardMap.NodesForShard(shard)
		if shard == shardToDrop {
			shardsToNodes[shard] = nodesExcept(nodes, srcNode)
		} else {
			shardsToNodes[shard] = nodes
		}
	}
	ts.UpdateShardMapping(shardsToNodes)
}

/*
 * Drop a shard from every node that hosts it, deliberately drop / lose data.
 */
func (ts *TestSetup) DropShard(shardToDrop int) {
	shardsToNodes := make(map[int][]string, ts.shardMap.NumShards())
	for shard := 1; shard <= ts.shardMap.NumShards(); shard++ {
		if shard == shardToDrop {
			shardsToNodes[shard] = make([]string, 0)
		} else {
			shardsToNodes[shard] = ts.shardMap.NodesForShard(shard)
		}
	}
	ts.UpdateShardMapping(shardsToNodes)
}

func (ts *TestSetup) DropRandomShard() {
	shardToDrop := rand.Int()%ts.NumShards() + 1
	ts.DropShard(shardToDrop)
}

func (ts *TestSetup) Shutdown() {
	for _, node := range ts.nodes {
		node.Shutdown()
	}
}

/*
 * Run test with a setup and shuts down the setup after
 */
func RunTestWith(t *testing.T, testFunc func(*testing.T, *TestSetup), setup *TestSetup) {
	testFunc(t, setup)
	setup.Shutdown()
}

/*
 * Makes a set of N nodes labeled n1, n2, n3, ... with fake Address/Port.
 */
func makeNodeInfos(n int) map[string]kv.NodeInfo {
	nodes := make(map[string]kv.NodeInfo)
	for i := 1; i <= n; i++ {
		nodes[fmt.Sprintf("n%d", i)] = kv.NodeInfo{Address: "", Port: int32(i)}
	}
	return nodes
}

func MakeBasicOneShard() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 1,
		Nodes:     makeNodeInfos(1),
		ShardsToNodes: map[int][]string{
			1: {"n1"},
		},
	}
}

func MakeMultiShardSingleNode() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 5,
		Nodes:     makeNodeInfos(1),
		ShardsToNodes: map[int][]string{
			1: {"n1"},
			2: {"n1"},
			3: {"n1"},
			4: {"n1"},
			5: {"n1"},
		},
	}
}

func MakeNoShardAssigned() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards:     1,
		Nodes:         makeNodeInfos(1),
		ShardsToNodes: map[int][]string{},
	}
}

func MakeSingleNodeHalfShardsAssigned() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 8,
		Nodes:     makeNodeInfos(1),
		ShardsToNodes: map[int][]string{
			1: {"n1"},
			2: {"n1"},
			3: {"n1"},
			4: {"n1"},
			5: {},
			6: {},
			7: {},
			8: {},
		},
	}
}

func MakeTwoNodeBothAssignedSingleShard() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 1,
		Nodes:     makeNodeInfos(2),
		ShardsToNodes: map[int][]string{
			1: {"n1", "n2"},
		},
	}

}

func MakeTwoNodeMultiShard() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 10,
		Nodes:     makeNodeInfos(2),
		ShardsToNodes: map[int][]string{
			1:  {"n1"},
			2:  {"n1"},
			3:  {"n1"},
			4:  {"n1"},
			5:  {"n1"},
			6:  {"n2"},
			7:  {"n2"},
			8:  {"n2"},
			9:  {"n2"},
			10: {"n2"},
		},
	}
}

func MakeFourNodesWithFiveShards() kv.ShardMapState {
	return kv.ShardMapState{
		NumShards: 5,
		Nodes:     makeNodeInfos(4),
		ShardsToNodes: map[int][]string{
			1: {"n1", "n2", "n3"},
			2: {"n1", "n2", "n4"},
			3: {"n2", "n3"},
			4: {"n3", "n4"},
			5: {"n2"},
		},
	}
}

func MakeManyNodesWithManyShards(numShards int, numNodes int) kv.ShardMapState {
	shardsToNodes := make(map[int][]string, numShards)
	nodeNames := make([]string, 0, numNodes)
	for i := 1; i <= numNodes; i++ {
		nodeNames = append(nodeNames, fmt.Sprintf("n%d", i))
	}
	for shard := 1; shard <= numShards; shard++ {
		// pick r unique nodes
		minR := 2
		maxR := 5
		r := rand.Intn(maxR-minR) + minR

		replicas := make(map[string]struct{}, 0)
		nodes := make([]string, 0)
		for {
			node := nodeNames[rand.Intn(numNodes)]
			replicas[node] = struct{}{}
			nodes = append(nodes, node)
			if len(replicas) >= r {
				break
			}
		}
		shardsToNodes[shard] = nodes
	}
	return kv.ShardMapState{
		NumShards:     numShards,
		Nodes:         makeNodeInfos(numNodes),
		ShardsToNodes: shardsToNodes,
	}
}

func randomString(rng *rand.Rand, length int) string {
	chars := "abcdefghijklmnopqrstuvwxyz"

	out := strings.Builder{}
	for i := 0; i < length; i++ {
		out.WriteByte(chars[rng.Int()%len(chars)])
	}
	return out.String()
}

func RandomKeys(n, length int) []string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	out := make([]string, 0)
	for i := 0; i < n; i++ {
		out = append(out, randomString(rng, length))
	}
	return out
}
