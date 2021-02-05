Write with Websocket
--------------------

Write records using Websocket protocol.

**GET** `/logs/{name}/records`  

Upgrade: websocket  
Connection: Upgrade  
X-HTTP-Method-Override: POST

### Params 

| Name           	| In     	| Description                                                     	| Default                    	|
|----------------	|--------	|-----------------------------------------------------------------	|----------------------------	|
| `name`         	| path   	| Log name.                                                       	|                            	|

### Response 

```
Status: 101 Switching protocol
```

### Code samples

**Wsdump** (_Requires [websocket-client](https://pypi.org/project/websocket-client-py3/) package._)
```bash
$ echo 'my record content' | wsdump.py ws://localhost:8000/logs/myLog/records --headers 'X-HTTP-Method-Override: POST'
```

**Python** (_Requires [websocket-client](https://pypi.org/project/websocket-client-py3/) package._)

```python
endpoint = 'ws://localhost:8000/logs/myLog/records'

ws = websocket.create_connection(endpoint, header=['X-HTTP-Method-Override: POST'])

record = 'my record content'

for i in range(10):
  ws.send(record)
```

**Go** (_Requires [github.com/gorilla/websocket](http://github.com/gorilla/websocket) package._)

```golang
  dialer := websocket.Dialer{}

  endpoint := "ws://localhost:8000/logs/myLog/records"

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