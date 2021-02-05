Read using HTTP
---------------

Read records using HTTP protocol.

**GET** `/logs/{name}/records`  

### Params 

| Name             	| In     	| Description                                                                                                                  	| Default                    	|
|------------------	|--------	|------------------------------------------------------------------------------------------------------------------------------	|----------------------------	|
| `name`           	| path   	| Log name.                                                                                                                    	|                            	|
| `whence`         	| query  	| Allowed values are `origin`, `start` and `end`.                                                                              	| `origin`                   	|
| `position`       	| query  	| Whence relative position from which the records are read from.                                                               	| `0`                        	|
| `count`          	| query  	| Limits the number of records to read, `-1` means no limitation.<br>Not available with `application/octet-stream` media type. 	| `-1`                       	|
| `follow`         	| query  	| Read will block until new records are written to the log.<br>Not available with `application/octet-stream` media type.       	| `false`                    	|
| `Accept`         	| header 	| See [Media-Types](/docs/api/media_types.md) for allowed values.                                                              	| `application/octet-stream` 	|
| `X-Styx-Timeout` 	| header 	| Number of seconds before timing out when waiting for new records with the `follow` query param.                              	|                            	|

### Response 

```
Status: 200 OK
```

Response contains records formatted according to `Accept`header.  

### Codes samples

#### Read the first available record

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start'

  res = requests.get(endpoint)
```

**Go**

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?whence=start"

  res, err := client.Get(endpoint)
  if err != nil {
    log.Fatal(err)
  }
```

#### Read first ten available records.

**Python** (_Requires [requests](https://pypi.org/project/requests/) package._)

```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start&count=10'

  headers = {
    'Accept': 'application/ld+text;line-ending=lf'
  }
  res = requests.get(endpoint, headers=headers)
```

**Go**

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