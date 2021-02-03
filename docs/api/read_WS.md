Read with Websocket
-------------------

### Read with websocket

Read all records from `myLog` and wait for more to be written using websocket.

Python:
```python
endpoint = 'ws://localhost:8000/logs/myLog/records'
ws = websocket.create_connection(endpoint)

while True:
  record = ws.recv()
  print(record.decode('utf_8'))
```

Go:
```golang
  dialer := websocket.Dialer{}

  endpoint := "ws://localhost:8000/logs/myLog/records?whence=start&position=0"

  headers := http.Header{}
  headers.Set("Origin", "localhost")

  conn, res, err := dialer.Dial(endpoint, headers)
  if err != nil {
    log.Fatal(err)
  }

  if res.StatusCode != http.StatusSwitchingProtocols {
    log.Fatal("an error occured")
  }

  defer conn.Close()

  for {
      _, record, err := conn.ReadMessage()
      if err != nil {
          log.Fatal(err)
      }

      fmt.Println(string(record))
  }

```
