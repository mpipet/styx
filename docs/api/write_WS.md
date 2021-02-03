Write with Websocket
--------------------

### Write records with websocket

Write ten records to `myLog` using websocket.

Python:
```python
endpoint = 'ws://localhost:8000/logs/myLog/records'

ws = websocket.create_connection(endpoint, header=['X-HTTP-Method-Override: POST'])

record = 'my record content'

for i in range(10):
  ws.send(record)
```

Go:
```golang
  dialer := websocket.Dialer{}

  endpoint := "ws://localhost:8000/logs/myLog/records?whence=start&position=0"

  headers := http.Header{}
  headers.Set("Origin", "localhost")
  headers.Set("X-HTTP-Method-Override", "POST")

  conn, resp, err := dialer.Dial(endpoint, headers)
  if err != nil {
    log.Fatal(err)
  }

  if resp.StatusCode != http.StatusSwitchingProtocols {
    log.Fatal("an error occured")
  }

  defer conn.Close()

  record := []byte("my record content")

  for i := 0; i < 10; i++ {
      err = conn.WriteMessage(websocket.BinaryMessage, record)
      if err != nil {
          log.Fatal(err)
      }
  }
```