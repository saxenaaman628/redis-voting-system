# redis-voting-system

## Background
In a world driven by interactivity and instant feedback, voting systems are no longer limited to political elections. From reality TV shows to product feedback campaigns, from live event polling to employee surveys — platforms today demand:

* Real-time vote aggregation
* Concurrent user handling (hundreds of thousands to millions)
* Live analytics dashboards
* Fraud prevention mechanisms
* Flexibility to scale up/down instantly

Traditional relational databases struggle under this pressure, often leading to poor user experiences, slow updates, and even system crashes. We sought a better solution.

## The Opportunity: Redis 8 — Beyond the Cache
Redis is widely known as a caching system, but Redis 8 unlocks the full power of:
* **Streams** for real-time logs and data ingestion
* **Sorted Sets** for ranking vote options instantly
* **Pub/Sub** for pushing updates to frontend/UI in real-time
* **Access Control Lists (ACL)** for secure, role-based user handling
* **RedisAI (Optional)** for integrating smart models to flag abnormal patterns
This makes Redis not just fast—but **functionally deep** enough to power a **primary database and logic system** for voting.

## The Mission
To build a **real-time voting platform** that is:
* **Configurable**: Any organization can set up an election/poll with custom options
* **Live**: Instant updates as users vote
* **Smart**: AI-based fraud detection module
* **Tested**: Simulates real-world concurrency (up to 1 million users voting simultaneously)
* **Secure**: Supports user authentication & authorization with rate-limiting

## Core System Design Principles
* **Idempotency**: Each user can only vote once; enforced using unique keys with TTL
* **Real-Time Data Ingestion**: Redis Streams log every action with metadata
* **High Concurrency Simulation**: We use Golang to simulate 1M parallel dummy users for load testing
* **Real-Time UI Sync**: Pub/Sub system updates frontend interfaces instantly
* **Auditability**: Streams and logs can be replayed or queried for analytics
* **Pluggable AI**: RedisAI models can analyze incoming votes for anomalies, bot-like patterns, etc.

## Simulation Focus (Upcoming)
To validate Redis as a reliable backbone, we are designing stress tests:
* **Golang-based concurrent simulation** of 1 million users casting votes
* **Monitoring system performance**: latency, throughput, stream growth, Pub/Sub responsiveness
* **Validating security**: blocked repeated votes, ACL-based user permissions

## Work-in-Progress
We’re currently in the development and architecture phase. Our core principles are flexibility, speed, and integrity. This project aims to redefine what real-time voting platforms can do, powered by Redis as a primary system—not just a side cache.

## Architecture Diagram and Components
<img width="700" height="650" alt="image" src="https://github.com/user-attachments/assets/6135f072-fcbc-4629-a18a-100c87312c30" />
