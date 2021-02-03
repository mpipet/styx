Read using HTTP
---------------

### Read a record

Read the first available record from `myLog`.

Python:
```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start'

  res = requests.get(endpoint)

  print(res.text)
```

Go:
```golang
  endpoint := "http://localhost:8000/logs/myLog/records?whence=start"

  client := &http.Client{}

  res, err := client.Get(endpoint)
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()

  if res.StatusCode != http.StatusOK {
    log.Fatal("an error occured")
  }

  record, err := ioutil.ReadAll(res.Body)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(string(record))
```

### Read line delimited records

Read first ten available records from `myLog` using line delimitation.

Python:
```python
  endpoint = 'http://localhost:8000/logs/myLog/records?whence=start&count=10'

  headers = {
    'Accept': 'application/ld+text;line-ending=lf'
  }
  res = requests.get(endpoint, headers=headers)

  sys.stdout.write(res.text)
```

Go:
```golang
  endpoint := "http://localhost:8000/logs/myLog/records?whence=start&count=10"

  client := &http.Client{}

  req, err := http.NewRequest(http.MethodGet, endpoint, nil)
  if err != nil {
    log.Fatal(err)
  }

  req.Header.Add("Accept", "application/ld+text;line-ending=lf")

  res, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
  defer res.Body.Close()

  if res.StatusCode != http.StatusOK {
    log.Fatal("an error occured")
  }

  _, err = io.Copy(os.Stdout, res.Body)
  if err != nil {
    log.Fatal(err)
  }
```