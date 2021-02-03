Read with Websocket
-------------------

### Read with websocket

Read all records from `myLog` and wait for more to be written using websocket.

Python

_Requires `websocket` package._

```python
endpoint = 'ws://localhost:8000/logs/myLog/records'
ws = websocket.create_connection(endpoint)

while True:
  record = ws.recv()
```

Go

_Requires `github.com/gorilla/websocket` package._

```golang
  dialer := websocket.Dialer{}

  endpoint := "ws://localhost:8000/logs/myLog/records?whence=start&position=0"

  headers := http.Header{}
  headers.Set("Origin", "localhost")

  conn, res, err := dialer.Dial(endpoint, headers)
  if err != nil {
    log.Fatal(err)
  }

  for {
      _, record, err := conn.ReadMessage()
      if err != nil {
          log.Fatal(err)
      }
  }
```
