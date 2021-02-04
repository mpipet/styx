HTTP long polling
----------------------

Styx HTTP api provides means to read log records using long polling.

_The following Python examples require `requests` package._

Assuming you have never consumed any records from the `myLog` and you want to retrieves its content in batchs of 100 line delimited records.

Python

```python
endpoint = 'http://localhost:8000/logs/myLog/records?count=100'

headers = {
  'Accept': 'application/ld+text;line-ending=lf'
}
res = requests.get(endpoint, headers=headers)
```

Go

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?count=100"

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

Assuming you effectively read 100 records, you have to increment the `position` query params with the number of processed records to consume the next ones.

Python

```python
endpoint = 'http://localhost:8000/logs/myLog/records?position=100&count=100'

headers = {
  'Accept': 'application/ld+text;line-ending=lf'
}
res = requests.get(endpoint, headers=headers)
```

Go

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?position=100&count=100"

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

If there was only 50 records to read, a response will be returned with the remaining records.
Now you want to wait for new records to be written by adding the `follow` query param.

Python

```python
endpoint = 'http://localhost:8000/logs/myLog/records?position=150&count=100&follow=true'

headers = {
  'Accept': 'application/ld+text;line-ending=lf'
  'X-Styx-Timeout': '30',
}
res = requests.get(endpoint, headers=headers)
```

Go

```golang
  client := &http.Client{}

  endpoint := "http://localhost:8000/logs/myLog/records?position=150&count=100&follow=true"

  req, err := http.NewRequest(http.MethodGet, endpoint, nil)
  if err != nil {
    log.Fatal(err)
  }

  req.Header.Add("Accept", "application/ld+text;line-ending=lf")
  req.Header.Add("X-Styx-Timeout", "30")

  res, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
```

The request will hang, up to the number of seconds specified by `X-Styx-Timeout` header, or until new records are written to the log.
As soon as new records are available these will be returned in the response, closing the HTTP communication.

With an infinite loop you can easily create a long running consummer using long polling.   
Here is the full code example.

Pyhon

```python
  headers = {
    'Accept': 'application/ld+text;line-ending=lf',
    'X-Styx-Timeout': '30'
  }

  position = 0

  while True:
    endpoint = 'http://localhost:8000/logs/myLog/records?position='+ str(position) +'&count=100&follow=true'

    res = requests.get(endpoint, headers=headers)

    for line in res.iter_lines():
      print(line.decode('utf-8'))
      
      position += 1
```

Go

```golang
  client := &http.Client{}

  position := int64(0)

  for {
    endpoint := "http://localhost:8000/logs/myLog/records?position=" + strconv.FormatInt(position, 10) + "&count=100&follow=true"

    req, err := http.NewRequest(http.MethodGet, endpoint, nil)
    if err != nil {
      log.Fatal(err)
    }

    req.Header.Set("Accept", "application/ld+text;line-ending=lf")
    req.Header.Set("X-Styx-Timeout", "30")

    res, err := client.Do(req)
    if err != nil {
      log.Fatal(err)
    }

    if res.StatusCode != http.StatusOK {
      continue
    }

    scanner := bufio.NewScanner(res.Body)
    scanner.Split(bufio.ScanLines)

    for scanner.Scan() {
      fmt.Println(scanner.Text())

      position += 1
    }

    res.Body.Close()
  }
```