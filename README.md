<img src="https://gitlab.com/dataptive/styx/-/raw/master/docs/logo.png" alt="Styx" width="250">

## What is Styx ?

Styx is a simple, high-performance event streaming database.

In essence, Styx provides a simple REST API to a very fast commit log engine. The system is built around the concepts of a log and of log records. The logs are append-only and immutable, and records can be produced and consumed to and from logs in a streaming fashion. Both logs and records are exposed as REST resources and can be accessed in multiple ways, from simple HTTP request-response cycles to a high-performance asynchronous binary procotol when raw performance is critical.

Styx was built to scratch our own itch, realizing the shortcomings of existing solutions, with the goal of making event streaming architectures within reach of most developers and companies. We started this project as an experiment to see how simple we could make event streaming, and quickly got caught in searching how we could make it market competitive performance and security-wise. Turns out Styx is able to handle millions of fsynced events per second at GB/s throughputs on a single node, on commodity hardware.

More specifically we tried to address:

- **Ease of integration and language compatibility**: produce and consume streams in a few lines of code from any language equipped with a HTTP or Websockets client. Producing and consuming can be as simple as a GET or a POST request. Produce and consume in batches if needed. GET supports long polling.
- **Ease of operation**: single binary and no dependencies to manage, full REST API, CLI, simple TOML configuration, out-of-the box Prometheus and Statsd monitoring, simple backup and restore, multiple retention options, runs well in docker.
- **Data safety**: records are immutable, atomic, durable, and fsynced to permanent storage before being acked. Styx detects data corruption at the record level in case of storage failure.
- **Performance**: very fast binary protocol available as an HTTP upgrade when performance is critical. Millions of events per second on a single node, GB/s thoughput, low latency. Scales to thousands of concurrent producers and consumers.

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

## Getting started

These examples assume you've built from source and got a running Styx server. You may need to adapt the commands if you run from the docker image.

### Basic log management

List available logs

```bash
./styx logs list
```

Create a test log

```bash
./styx logs create test
```

Follow the log contents from a terminal

```bash
./styx logs read test --follow
```

Write to the log from another terminal

```bash
./styx logs write test
< hello
< styx !
```

You should see the records reflected on the reading side.

List the available logs again. You should see your log stats updated.

```bash
./styx logs list
```

Delete the test log

```bash
./styx logs delete test
```

The CLI exposes the full log management API. Use the `--help` flag to get a summary of available commands.

```bash
./styx logs --help
```

To get detailed help on a command use the `--help` flag after the command name. For example:

```bash
./styx logs create --help
```

### Consuming logs

Create an empty log

```bash
./styx logs create test
```

Fill it with a few lines

```bash
./styx logs write test
< styx
< is
< really
< awesome !
```

You can read it from beggining to end

```bash
./styx logs read test
```

Or read only the first two records

```bash
./styx logs read test -n 2
```

Or read the last two records by starting from the record at position 2.

```bash
./styx logs read test --position=2
```

You can accomplish the same by using relative positioning. Here we request position -2 from the end of the log.

```bash
./styx logs read test --whence=end --position=-2
```

To follow a log in real-time, use the `--follow` or `-f` flag.

```bash
./styx logs read test -f
```

If you simultaneously write to the log you'll get the updates in real time.

## Documentation

For further information see the [Documentation](/docs), the [API Reference](/docs/api), the [Examples](/docs/examples) or the [Howto's](/docs/howto).

## Licence

Styx is published under a permissive Apache 2.0 + BSL licence. This is not legal advice and cannot be sustituted to the [LICENSE](LICENSE) file, but it means in short that unless you are a cloud provider or plan to make money by selling styx as a service, you're safe to use it for any purpose you see fit.

This project is source-available for now, but as open-source lovers we plan to open it progressively. We choosed Apache 2.0 + BSL after much thought and debate to be able to build a sustainable business that will ultimately feed into Styx development and features. We did not multi-license the repository to keep things simple for now, but we plan on relaxing licensing on the most code surface possible as early as possible. We prefer to start with restrictions (though minimal) and remove them along the way, than going the other way around and deceiving users and contributors.

## Contributions

We do not accept code contributions for now for stability reasons. In particular, patching the core requires deep knowledge of the architecture and extensive testing. Reviewing submissions may be difficult to scale for the two of us, so we prefer to keep it closed for now.

That being said, you are more than wecome to post issues, bug reports, or feature suggestions. We are open to discussion and eager to learn about your use cases !

## Community

You can join our Slack server at [...].



