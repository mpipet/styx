Styx
====

Creating a 1 node cluster
-------------------------

Setup data directory

```bash
mkdir data
mkdir -p ./data/node1/logs
mkdir -p ./data/node1/raft
sed "s/NUM/1/" < config.toml > ./data/node1/config.toml
```

Run node

```bash
go run cmd/styx-server/main.go --config ./data/node1/config.toml --log-level TRACE
```

Make node a leader of it own cluster

```bash
curl localhost:8001/nodes/bootstrap -X POST
```

Check node state


Creating a 3 node cluster
-------------------------

Setup data directory

```bash
mkdir data
for NUM in 1 2 3
do
	mkdir -p ./data/node$NUM/logs
	mkdir -p ./data/node$NUM/raft
    sed "s/NUM/$NUM/" < config.toml > ./data/node$NUM/config.toml
done
```

Run nodes (in separate terminals)

```bash
go run cmd/styx-server/main.go --config ./data/node1/config.toml --log-level TRACE
go run cmd/styx-server/main.go --config ./data/node2/config.toml --log-level TRACE
go run cmd/styx-server/main.go --config ./data/node3/config.toml --log-level TRACE
```

Setup cluster

```bash
curl localhost:8001/nodes/bootstrap -X POST
curl localhost:8001/nodes -X POST -d name=node2 -d address=127.0.0.1:8002
curl localhost:8001/nodes -X POST -d name=node3 -d address=127.0.0.1:8003
```

Check cluster state

```bash
curl localhost:8001/nodes
```
