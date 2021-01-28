<img src="https://gitlab.com/dataptive/styx/-/raw/master/docs/logo-styx.png" alt="Styx">


What is Styx ?
--------------

Styx is a simple, high-performance event streaming database.

In essence, Styx provides a simple REST API to a very fast commit log engine. The system is built around the concepts of a log and of log records. The logs are append-only and immutable, and records can be produced and consumed to and from logs in a streaming fashion. Both logs and records are exposed as REST resources and can be accessed in multiple ways, from simple HTTP request-response cycles to a high-performance asynchronous binary procotol when raw performance is critical.

Styx was built to scratch our own itch, realizing the shortcomings of existing solutions, with the goal of making event streaming architectures within reach of most developers and companies. We started this project as an experiment to see how simple we could make event streaming, and quickly get caught in searching how we could make it market competitive performance-wise. Turns out Styx is able to handle millions of events per second at GB/s throughputs on a single node, on commodity hardware.

More specifically we tried to address :

- **Ease of integration and language compatibility**: produce and consume streams in a few lines of code from any language equipped with a HTTP client.
- **Ease of operation**: single binary and no dependencies thanks to Golang, full REST API, fully-featured CLI, simple configuration, out-of-the box Prometheus and Statsd monitoring.
- **Data safety**: records are immutable, atomic, durable, and fsynced to permanent storage before ack.
- **Performance**: millions of events per second on a single node, GB/s thoughput, low latency.

Features
--------

- Immediately integrated: produce and consume single or batched events through GET and POST requests
- Integration options: HTTP long-polling, WebSockets, simple and fast binary protocol when every bit of performance is critical
- Fast data ingestion: directly push large line-delimited files with cURL
- Simple to manage: REST API, full-featured CLI
- Simple to deploy: one Golang binary, no dependencies
- Monitoring: out-of-the-box Prometheus, Statsd
- Backuping: fast and simple backup and restore through cURL or the styx CLI
- Strong guarantees: records written to logs are atomic and durable. Records acknowledged only after having been fsynced to permanent storage.
- Low latency: in the order of milliseconds.
- Seamless performance: automated server-side batching, so the client doesn't have to carry the complexity.
- Data security: data corruption detection at the record level.
- Simple configuration: simple TOML configuration with sane defaults.
- Data management: count, size and age based expiration policies.
- Fast access: sparse indexing for low latency seeks without impacting write performance.
- Scaling logs: storage engine designed to maintain sequential writing performance when pushing massive streams to multiple logs at once. 
- Scaling clients: handles thousands of producers and consumers at once.


Design
------



Running Styx
------------

Setup data directory

```bash
mkdir data
```

Run styx

```bash
go run cmd/styx-server/main.go --config ./config.toml --log-level TRACE
```