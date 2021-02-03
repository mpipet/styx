Read using HTTP
---------------

### Read a record

Read the first available record from `myLog`.

Python

_Requires `requests` package._
```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start'

  res = requests.get(endpoint)

  print(res.text)
```

Go

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?whence=start"

  res, err := client.Get(endpoint)
  if err != nil {
    log.Fatal(err)
  }
```

Available query parameters

| Param      | Description                                                                                                                                                                                                | Default  |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `whence`   | Allowed values are:<br>`origin` the equivalent to the start position of the log at creation time.<br>`start` the first available position.<br>`end` the last available position.                           | `origin` |
| `position` | Whence relative position from wich the records are read from.<br>Negative values are allowed.<br>Example: With `whence=end` and `position=-10`, read will start from 10 records before the end of the log. | `0`      |


### Read line delimited records

Read first ten available records from `myLog` using line delimitation.

Python

_Requires `requests` package._
```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start&count=10'

  headers = {
    'Accept': 'application/ld+text;line-ending=lf'
  }
  res = requests.get(endpoint, headers=headers)
```

Go

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?whence=start&count=10"

  req, err := http.NewRequest(http.MethodGet, endpoint, nil)
  if err != nil {
    log.Fatal(err)
  }

  req.Header.Add("Accept", "application/ld+text;line-ending=lf")

  res, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
```

Available query parameters

| Param      | Description                                                                                                                                                                                                | Default  |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `whence`   | Allowed values are:<br>`origin` the equivalent to the start position of the log at creation time.<br>`start` the first available position.<br>`end` the last available position.                           | `origin` |
| `position` | Whence relative position from wich the records are read from.<br>Negative values are allowed.<br>Example: With `whence=end` and `position=-10`, read will start from 10 records before the end of the log. | `0`      |
| `count`    | Limits the number of records to read, `-1` means no limitation.                                                                                                                                            | `-1`     |
| `follow`   | Read will block until new records are written to the log.<br>Since waiting forever is an unwanted behavior when using HTTP, `X-Styx-Timeout` header should be set.                                         | `false`  |


See [Media-Types](/docs/api/media_types.md) for details about `application/ld+text`.