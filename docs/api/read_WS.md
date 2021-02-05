Read with Websocket
-------------------

Read records using websocket protocol.

**GET** `/logs/{name}/records`

Upgrade: websocket  
Connection: Upgrade  

### Params 

| Name       	| In    	| Description                                                    	| Default  	|
|------------	|-------	|----------------------------------------------------------------	|----------	|
| `name`     	| path  	| Log name.                                                      	|          	|
| `whence`   	| query 	| Allowed values are `origin`, `start` and `end`.                	| `origin` 	|
| `position` 	| query 	| Whence relative position from which the records are read from. 	| `0`      	|

### Response 

```
Status: 101 Switching protocol
```

### Code samples

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
endpoint = 'ws://localhost:8000/logs/myLog/records'
ws = websocket.create_connection(endpoint)

while True:
  record = ws.recv()
```

**Go** (_Requires [github.com/gorilla/websocket](http://github.com/gorilla/websocket) package._)

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
