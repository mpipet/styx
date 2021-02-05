Write with HTTP
---------------

Write records using HTTP protocol.

**POST** `/logs/{name}/records`  

### Params

| Name           	| In     	| Description                                                     	| Default                    	|
|----------------	|--------	|-----------------------------------------------------------------	|----------------------------	|
| `name`         	| path   	| Log name.                                                       	|                            	|
| `Content-Type` 	| header 	| See [Media-Types](/docs/api/media_types.md) for allowed values. 	| `application/octet-stream` 	|

### Response 

```
Status: 200 OK
```
```json
{
  "position": 20,
  "count": 10,
}
```

### Codes samples

#### Write a record

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
  endpoint = 'http://localhost:8000/logs/myLog/records'

  record = bytes('my record content', 'utf-8')

  res = requests.post(endpoint, data=record)
```

**Go**

```golang
  endpoint := "http://localhost:8000/logs/myLog/records"

  client := &http.Client{}

  record := []byte("my record content")

  res, err := client.Post(endpoint, "application/octet-stream", bytes.NewReader(record))
  if err != nil {
    log.Fatal(err)
  }
```

#### Write line delimited records

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
  records = b''
  for i in range(10):
    records += bytes('my record content\n', 'utf-8')

  endpoint = 'http://localhost:8000/logs/myLog/records'

  headers = {
    'Content-Type': 'application/ld+text;line-ending=lf'
  }

  requests.post(endpoint, headers=headers, data=records)
```

**Go**

```golang
  client := &http.Client{}

  buf := bytes.NewBuffer([]byte{})
  record := []byte("my record content\n")

  for i := 0; i < 10; i++ {
    buf.Write(record)
  }

  endpoint := "http://localhost:8000/logs/myLog/records"

  res, err := client.Post(endpoint, "application/ld+text;line-ending=lf", buf)
  if err != nil {
    log.Fatal(err)
  }
```
