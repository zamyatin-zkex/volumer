# Volumer

**Volumer** is a real-time service for calculating trading volume statistics over different time intervals (5m, 1h, 24h, ...).

## What for

This project is a test assignment:

**Context**<br>
You have 1000 swaps per second coming from a producer (who, token, amount, usd, side,..). Producer also persists this data in db . Need to build a system that calculates real-time token statistics (5min volume, 1H volume, .. , 24h volume, transaction counts, etc) and serves this data via HTTP API and WebSocket updates with minimal latency. System must be highly available and handle restarts without losing data or missing events during startup time. It should be scalable, so we can spin more instances. Swaps data may contain duplicates and block order is not guaranteed.

**Theoretical**<br>
Design the complete architecture. What transport mechanisms would you use from producer? Where would you store different types of data? How would you ensure high availability and zero data loss?

**Practical**<br>
Implement the Go service that reads swap events from a channel, calculates statistics, serves the data over HTTP,  submits updates to a WebSocket channel and handles restarts. Use interfaces for storage. (edited)

## Architecture

Volumer consists of the following services:

- **faketrader** — generates trades and publishes them to Kafka.
- **consumer** — reads trades from Kafka and commits offsets when state is saved.
- **roller** — aggregates trading volume over specified time intervals.
- **watcher** — polls `roller` and updates data for `web`.
- **interrupter** — listens for system calls (syscalls).
- **web** — serves trading statistics via WebSockets and HTTP.

## Data Flow (Mermaid)

```mermaid
flowchart TD
    F[faketrader] -- "trades" --> K[Kafka]
    K -- "trades" --> C[consumer]
    C -- "processed trades" --> R[roller]
    R -- "request commit" --> C
    C -- "commit offsets" --> K
    R -- "volume stats" --> Wt[watcher]
    Wt -- "updated stats" --> W[web]
    W -- "HTTP / WebSocket responses" --> U[(Users)]
```

```mermaid
    sequenceDiagram
    participant F as faketrader
    participant K as Kafka
    participant C as consumer
    participant R as roller
    participant Wt as watcher
    participant W as web
    participant U as user

    F->>K: publish trades
    K->>C: consume trades
    C->>R: send trades
    R->>K: save current state to rolls topic
    R->>C: commit offset on trades topic
    C->>K: commit offsets
    R->>Wt: provide volume stats
    Wt->>W: update aggregated stats
    W->>U: serve stats (HTTP / WebSockets)
```

## Run
server:
```
make kafka
make run
```
client:
```
make http
make wsclient
```

## Known issues
- sarama SyncProducer is too slow, use AsyncProducer instead
- potential bottleneck with one topic partition
