Write with HTTP
---------------

### Write a record

Write one record to `myLog`.

Python

_Requires `requests` package._

```python
  endpoint = 'http://localhost:8000/logs/myLog/records'

  record = bytes('my record content', 'utf-8')

  res = requests.post(endpoint, data=record)
```

Go

```golang
  endpoint := "http://localhost:8000/logs/myLog/records"

  client := &http.Client{}

  record := []byte("my record content")

  res, err := client.Post(endpoint, "application/octet-stream", bytes.NewReader(record))
  if err != nil {
    log.Fatal(err)
  }
```

### Write line delimited records

Write ten line delimited records to `myLog`.

Python

_Requires `requests` package._

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

Go

```golang
  client := &http.Client{}

  buf := bytes.NewBuffer([]byte{})

  for i := 0; i < 10; i++ {
    record := []byte("my record content\n")
    buf.Write(record)
  }

  endpoint := "http://localhost:8000/logs/myLog/records"

  res, err := client.Post(endpoint, "application/ld+text;line-ending=lf", buf)
  if err != nil {
    log.Fatal(err)
  }
```
