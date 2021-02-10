<img src="https://gitlab.com/dataptive/styx/-/raw/master/docs/logo.png" alt="Styx" width="250">

## What is Styx ?

Styx is a simple, high-performance event streaming database.

In essence, Styx provides a simple REST API to a very fast commit log engine. The system is built around the concepts of a log and of log records. The logs are append-only and immutable, and records can be produced and consumed to and from logs in a streaming fashion. Both logs and records are exposed as REST resources and can be accessed in multiple ways, from simple HTTP request-response cycles to a high-performance asynchronous binary procotol when raw performance is critical.

Styx was built to scratch our own itch, realizing the shortcomings of existing solutions, with the goal of making event streaming architectures within reach of most developers and companies. We started this project as an experiment to see how simple we could make event streaming, and quickly got caught in searching how we could make it market competitive performance and security-wise. Turns out Styx is able to handle millions of fsynced events per second at GB/s throughputs on a single node, on commodity hardware.

More specifically we tried to address:

- **Ease of integration and language compatibility**: produce and consume streams in a few lines of code from any language equipped with a HTTP or Websockets client. Producing and consuming can be as simple as a GET or a POST request. Produce and consume in batches if needed. GET supports long polling.
- **Ease of operation**: single binary and no dependencies to manage, full REST API, CLI, simple TOML configuration, out-of-the box Prometheus and Statsd monitoring, simple backup and restore, multiple retention options, runs well in docker.
- **Data safety**: records are immutable, atomic, durable, and fsynced to permanent storage before being acked. Styx detects data corruption at the record level in case of storage failure.
- **Performance**: very fast binary protocol available as an HTTP upgrade when performance is critical. Millions of events per second on a single node, GB/s thoughput, low latency. Thousands of concurrent producers and consumers.

## Install

There are various ways of installing Styx.

### Using docker

Build the docker image

```bash
git clone https://gitlab.com/dataptive/styx.git
cd styx
docker build -t styx .
```

Run the docker image

```bash
docker run -it --rm -p 8000:8000 --name styx styx
```

Run the docker image with persistent data directory on the host

```bash
mkdir data
docker run -it --rm -p 8000:8000 --mount type=bind,source="$(pwd)"/data,target=/data --name styx styx
```

Execute a CLI command from the running container. In this instance you should see an empty log list.

```bash
docker exec -it styx styx logs list
```

### Building from source

Building from source requires golang and git installed on the target system.

Clone the repository

```bash
git clone https://gitlab.com/dataptive/styx.git
cd styx
```

Build the server

```bash
go build -o styx-server cmd/styx-server/main.go 
```

Build the CLI

```bash
go build -o styx cmd/styx/main.go 
```

Run the server. A default config file is provided at the root of the repository.

```bash
./styx-server --config=config.toml --log-level=TRACE
```

Check that the CLI can access the server. You should see an empty log list.

```bash
./styx logs list
```

## Documentation

See [Documentation](/docs) for more informations about Styx.
