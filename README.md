<img src="https://gitlab.com/dataptive/styx/-/raw/master/docs/logo.png" alt="Styx" width="250">

## What is Styx ?


Styx is a simple, high-performance event streaming database.

In essence, Styx provides a simple REST API to a very fast commit log engine. The system is built around the concepts of a log and of log records. The logs are append-only and immutable, and records can be produced and consumed to and from logs in a streaming fashion. Both logs and records are exposed as REST resources and can be accessed in multiple ways, from simple HTTP request-response cycles to a high-performance asynchronous binary procotol when raw performance is critical.

Styx was built to scratch our own itch, realizing the shortcomings of existing solutions, with the goal of making event streaming architectures within reach of most developers and companies. We started this project as an experiment to see how simple we could make event streaming, and quickly got caught in searching how we could make it market competitive performance and security-wise. Turns out Styx is able to handle millions of fsynced events per second at GB/s throughputs on a single node, on commodity hardware.

More specifically we tried to address :

- **Ease of integration and language compatibility**: produce and consume streams in a few lines of code from any language equipped with a HTTP or Websockets client. Producing and consuming can be as simple as a GET or a POST request. Produce and consume in batches if needed. GET supports long polling.
- **Ease of operation**: single binary and no dependencies to manage, full REST API, CLI, simple TOML configuration, out-of-the box Prometheus and Statsd monitoring, simple backup and restore, multiple retention options, runs well in docker.
- **Data safety**: records are immutable, atomic, durable, and fsynced to permanent storage before being acked. Styx detects data corruption at the record level in case of storage failure.
- **Performance**: Upgrade to Styx binary protocol when performance is critical. Millions of events per second on a single node, GB/s thoughput, low latency. Thousands of concurrent producers and consumers.

## Install

### Using docker

TODO

### Precompiled binaries

TODO

### Building from source

TODO

## Documentation

See [Documentation](/docs) for more informations about Styx.
