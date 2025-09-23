# Volumer

**Volumer** is a real-time service for calculating trading volume statistics over different time intervals (5m, 1h, 24h, ...).

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

    I[interrupter] -. "listens for syscalls" .-> I
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
    participant I as interrupter

    F->>K: publish trades
    K->>C: consume trades
    C->>R: send processed trades
    R->>C: request offset commit
    C->>K: commit offsets
    R->>Wt: provide volume stats on request
    Wt->>W: update aggregated stats
    W->>U: serve stats (HTTP / WebSockets)

    Note over I: listens for syscalls
```
