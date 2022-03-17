# [Preview] Lab 5: Default Final Project

You may design your own final project ([here](https://docs.google.com/document/d/1KAj0vWXFoL4ITRjF9h7PQu8xQiUOdg2P2Gqw0448V7A/edit?usp=sharing) is a list of potential ideas), but if none appeals to you can consider
this default lab. You must still submit a detailed proposal with your design as Lab 5 will not be spec'd out in details like Labs 0-4.

Detailed requirements on logistics can be found [here](https://docs.google.com/document/d/1KAj0vWXFoL4ITRjF9h7PQu8xQiUOdg2P2Gqw0448V7A/edit?usp=sharing).

## Quorum consistency for replicated stores

In lab 4 you implemented a sharded and replicated key-value store with essentially
no consistency guarantees -- if writes failed partially reads may or may not return
a result (or may return stale results). Replicas can diverge in unexpected ways,
and reads may return differing results based on which replica is chosen.

To solve this, you will implement a flexible quorum protocol, similar to Dynamo or
Cassandra. You may choose any conflict resolution strategy, including:
 - Last-write wins based on timestamps (or HLCs)
 - Vector clocks
 - Application driven merging
 - CRDTs, by changing the values stored
