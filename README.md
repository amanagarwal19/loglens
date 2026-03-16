# LogLens

A real-time log deduplication and semantic clustering pipeline built on Kafka and Go, with a React dashboard.

## The Problem

When something breaks in production, your logging tool floods you with thousands of repeated or near-identical log lines. You waste time manually filtering noise instead of investigating what is actually new and unique. LogLens solves this by grouping semantically similar errors together in real time.

## How It Works

1. Any service publishes structured log events to the `raw-logs` Kafka topic
2. The Go consumer reads each event and calls Voyage AI to generate an embedding vector for the log message
3. The cluster engine computes cosine similarity against existing cluster centroids
4. If similarity is above the threshold (default 0.90), the log is merged into that cluster, otherwise a new cluster is created
5. Clusters are persisted to Redis so they survive consumer restarts
6. The consumer exposes an HTTP API on port 4000 that the React dashboard polls every 3 seconds

## Architecture

```
Services → Kafka (raw-logs topic) → Go Consumer → Voyage AI Embeddings
                                         ↓
                                   Cluster Engine (cosine similarity)
                                         ↓
                                      Redis (persistence)
                                         ↓
                                   HTTP API (:4000)
                                         ↓
                                  React Dashboard (:5173)
```

## Project Structure

```
Go/
├── frontend/              # React dashboard (Vite)
│   └── src/
│       └── App.jsx        # Main dashboard component
└── loglens/               # Go backend
    ├── cluster/
    │   ├── cluster.go     # Cluster engine with cosine similarity
    │   └── store.go       # Redis persistence layer
    ├── consumer/
    │   └── consumer.go    # Kafka consumer + HTTP API server
    ├── embeddings/
    │   ├── embeddings.go  # Voyage AI embeddings via raw HTTP
    │   └── embeddings_test.go
    ├── models/
    │   ├── cluster.go     # Cluster struct
    │   └── events.go      # LogEvent struct
    ├── producer/
    │   └── producer.go    # Fake log producer for development
    └── docker-compose.yml # Kafka, Zookeeper, Redis
```

## Quick Start

### Prerequisites
- Go 1.22+
- Docker and Docker Compose
- Node.js 18+
- Voyage AI API key (free at voyageai.com)

### 1. Start infrastructure

```bash
cd loglens
docker compose up -d
```

### 2. Start the consumer

```bash
export VOYAGE_API_KEY=your-key-here
go run consumer/consumer.go
```

### 3. Start the producer

```bash
go run producer/producer.go
```

### 4. Start the dashboard

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173` in your browser.

## Key Concepts Learned

### Kafka

**Topics** are named, persistent logs. Unlike traditional queues, messages are not deleted when consumed. They are retained for a configurable period (default 7 days), enabling replay.

**Producers** append messages to topics. **Consumers** read from topics independently. They are fully decoupled and have no direct connection.

**Consumer groups** control how messages are distributed. Consumers in the same group split partitions between them for parallel processing. Consumers in different groups each receive all messages independently.

**Partitions** are the unit of parallelism. A topic with 3 partitions can have at most 3 consumers in a group working in parallel. Adding partitions increases throughput.

**Offsets** are bookmarks. Kafka tracks where each consumer group is up to. This enables replay: restart a consumer and it picks up exactly where it left off.

**Partition keys** route messages deterministically. Messages with the same key always go to the same partition, guaranteeing ordering per key.

**Rebalancing** happens automatically when consumers join or leave a group. Kafka redistributes partitions among active consumers, providing fault tolerance.

### Go

**Structs** define the shape of data. JSON struct tags (`json:"field_name"`) control serialization key names. `omitempty` skips fields that are empty.

**Interfaces** are satisfied implicitly. No `implements` keyword needed.

**Goroutines** are lightweight threads. The `go` keyword runs a function concurrently. Used to run the HTTP API server alongside the Kafka consumer loop.

**Channels** pass values between goroutines. Used for graceful shutdown by listening for OS signals.

**Defer** runs a statement when the surrounding function exits, regardless of how it exits. Used for cleanup like closing producers, consumers, and HTTP response bodies.

**Pointers** (`&`) pass memory addresses. Required when a function needs to modify a value you pass in. `json.Unmarshal` takes `&event` because it writes into the struct.

**Error handling** is explicit. Functions return `error` as the last return value. The caller always checks it. No exceptions.

**`sync.Mutex`** prevents race conditions when multiple goroutines access shared data. `Lock()` acquires exclusive access, `Unlock()` releases it. `defer mu.Unlock()` guarantees release even if the function panics.

**Modules** are declared in `go.mod`. Import paths follow the pattern `modulename/packagename`.

### Embeddings and Similarity

**Embeddings** convert text into vectors (lists of numbers). Semantically similar text produces vectors that are mathematically close to each other.

**Cosine similarity** measures the angle between two vectors. Returns a value between 0 and 1. 1 means identical direction (very similar), 0 means perpendicular (unrelated).

**Centroids** represent the average of all vectors in a cluster. Updated using a rolling average so you never need to store all individual vectors.

**Threshold tuning**: a similarity threshold of 0.90 means logs must be 90% similar to be merged. Higher values create more clusters (stricter matching), lower values create fewer clusters (looser matching).

### React

**Components** are functions that return JSX. JSX looks like HTML but compiles to JavaScript.

**useState** stores values that survive between renders. Returns `[value, setter]`. Calling the setter triggers a re-render with the new value.

**useEffect** runs side effects at specific times. The second argument controls when: `[]` means run once on mount, `[dep]` means run when `dep` changes.

**Cleanup functions** returned from `useEffect` run when the component unmounts. Used to clear intervals and prevent memory leaks.

**Props** pass data between components. Parent passes data down, child receives it as function arguments. Used for `setClusters` prop drilling from `App` to `Header` to `ResetClustersButton`.

**Keys** on list items help React identify which items changed. Always use a unique, stable identifier like an ID, never an array index.

**Data transformation** is often needed before passing data to third party libraries like Recharts. The API response shape rarely matches what a library expects exactly.

### HTTP and APIs

**CORS** (Cross Origin Resource Sharing) is a browser security rule that blocks requests from one origin to another unless the server explicitly allows it with `Access-Control-Allow-Origin: *`.

**Method routing** in Go 1.22+ uses `mux.HandleFunc("GET /path", handler)` syntax. The method is part of the route pattern.

**Raw HTTP calls** to external APIs require setting headers manually. The `Authorization: Bearer key` header is the standard for REST API authentication.

**`json.NewDecoder`** streams JSON from an `io.Reader` without loading the entire response into memory first. More efficient than `json.Unmarshal` for large responses.

**Response body cleanup** requires `defer resp.Body.Close()` after every HTTP call. Unclosed bodies leak network connections.

### Redis

**Key-value store** where everything is stored under a string key. `SET key value` stores, `GET key` retrieves, `KEYS pattern*` lists matching keys.

**TTL (time to live)** controls expiry. `0` means no expiry. In production you would set a TTL to automatically clean up stale data.

**Wildcard patterns** like `cluster:*` match all keys with that prefix. Used to load all clusters on startup and to reset them.

**Persistence** survives process restarts. Clusters built up over hours are not lost when you restart the consumer.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `VOYAGE_API_KEY` | required | Voyage AI API key for embeddings |
| Kafka broker | `localhost:9092` | Set in consumer and producer code |
| Redis address | `localhost:6379` | Set in consumer code |
| Consumer API port | `4000` | HTTP API for dashboard |
| Dashboard port | `5173` | Vite dev server |
| Similarity threshold | `0.90` | Controls cluster merging aggressiveness |

## Kafka Topic Setup

The `raw-logs` topic is auto-created by Kafka. To manually control partitions:

```bash
# delete existing topic
docker exec -it loglens-kafka-1 kafka-topics --delete \
  --topic raw-logs \
  --bootstrap-server localhost:9092

# recreate with 3 partitions
docker exec -it loglens-kafka-1 kafka-topics --create \
  --topic raw-logs \
  --partitions 3 \
  --replication-factor 1 \
  --bootstrap-server localhost:9092

# inspect topic
docker exec -it loglens-kafka-1 kafka-topics --describe \
  --topic raw-logs \
  --bootstrap-server localhost:9092
```

## Redis Commands

```bash
# view all clusters
docker exec -it loglens-redis-1 redis-cli KEYS "cluster:*"

# read a specific cluster
docker exec -it loglens-redis-1 redis-cli GET "cluster:your-id"

# wipe everything
docker exec -it loglens-redis-1 redis-cli FLUSHALL
```

## Next Steps

- [ ] Replace in-memory vector search with pgvector or Qdrant for scalable similarity search
- [ ] Add Kafka Streams for windowed aggregation (errors per minute per service)
- [ ] Slack webhook alerting when a new cluster is created
- [ ] Dead letter queue for malformed events
- [ ] Dockerfiles for each service
- [ ] Helm chart for Kubernetes deployment
- [ ] Authentication on the dashboard API
- [ ] Configurable similarity threshold via environment variable