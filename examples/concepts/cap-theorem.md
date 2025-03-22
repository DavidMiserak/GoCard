---
tags: [distributed-systems,cap-theorem,consistency,availability,partition-tolerance]
created: 2025-03-22
last_reviewed: 2025-03-22
review_interval: 0
difficulty: 0
---

# CAP Theorem

## Question

What is the CAP theorem, and how does it influence distributed system design? Give examples of systems making different CAP tradeoffs.

## Answer

The CAP theorem (also known as Brewer's theorem) states that a distributed data store cannot simultaneously provide more than two out of the following three guarantees:

### The Three Guarantees

1. **Consistency (C)**: Every read receives the most recent write or an error
2. **Availability (A)**: Every request receives a non-error response, without guarantee that it contains the most recent write
3. **Partition Tolerance (P)**: The system continues to operate despite network partitions (nodes cannot communicate)

### Key Insight

In a distributed system, network partitions are inevitable, so you must choose between consistency and availability when a partition occurs:

- **CP systems**: Prioritize consistency over availability during partitions
- **AP systems**: Prioritize availability over consistency during partitions
- **CA systems**: Cannot exist in real-world distributed systems (as partitions are unavoidable)

### Real-World System Examples

| Type | Examples | Characteristics |
|------|----------|-----------------|
| **CP** | - Google's Spanner<br>- HBase<br>- Apache ZooKeeper<br>- etcd | - Strong consistency<br>- May become unavailable during partitions<br>- Often use consensus protocols like Paxos/Raft |
| **AP** | - Amazon Dynamo<br>- Cassandra<br>- CouchDB<br>- Riak | - High availability<br>- Eventually consistent<br>- Often use techniques like vector clocks, gossip protocols |
| **CA** | - Single-node relational databases<br>- (Not truly distributed) | - Cannot tolerate partitions<br>- Useful comparison benchmark |

### Design Implications

1. **When to Choose CP**:
   - Financial systems requiring strong consistency
   - Systems where incorrect data is worse than no data
   - Configuration management, leader election, service discovery

2. **When to Choose AP**:
   - High-traffic web applications
   - Real-time analytics
   - Systems where stale data is acceptable
   - Content delivery networks

3. **Strategies to Mitigate Tradeoffs**:
   - PACELC theorem extension: When there's no partition, choose between latency and consistency
   - Tunable consistency levels (e.g., Cassandra's quorum reads)
   - CRDT (Conflict-free Replicated Data Types)
   - Event sourcing and Command Query Responsibility Segregation (CQRS)
