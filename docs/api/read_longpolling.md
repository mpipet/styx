Read using longpolling
----------------------

### Read with long polling

Read all records from `myLog` and wait for more to be written using HTTP long polling.

Python:
```python
  headers = {
    'Accept': 'application/ld+text;line-ending=lf',
    'X-Styx-Timeout': '30'
  }

  position = 0

  while True:
    endpoint = 'http://localhost:8000/logs/myLog/records?position='+ str(position) +'&follow=true'

    res = requests.get(endpoint, headers=headers)

    for line in res.iter_lines():
      print(line.decode('utf-8'))
      
      position += 1
```

Go:
```golang
  client := &http.Client{}

  position := int64(0)

  for {
    endpoint := "http://localhost:8000/logs/myLog/records?position=" + strconv.FormatInt(position, 10) + "&follow=true"

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
