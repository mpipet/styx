python-gdax-connector
=====================

Install requirements

```bash
pip3 install -r requirements.txt
```

Ensure Styx is running and create a "gdax" log.

```bash
curl localhost:8000/logs -X POST -d name=gdax
```

Run the connector

```bash
python3 main.py
```

From another terminal, check that the log is filling with GDAX events

```bash
curl localhost:8000/logs
```

Try consuming the events from another terminal.

```bash
https://github.com/websocket-client/websocket-client.git
cd websocket-client/bin
python3 wsdump.py
python3 wsdump.py ws://localhost:8000/logs/gdax/records
```

