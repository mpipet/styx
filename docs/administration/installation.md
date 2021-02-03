Install
-------

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

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

Running Styx with Docker
------------------------

Build:
```bash
docker build -t styx .
```

Run:
```bash
docker run -it --rm -p 8000:8000 --name styx styx
```

Run using host data directory:
```bash
docker run -it --rm -p 8000:8000 --mount type=bind,source="$(pwd)"/data,target=/data --name styx styx
```