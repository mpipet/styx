Read with Styx protocol
------------------------

Read records using [Styx protocol](/docs/api/styx_protocol.md).

**GET** `/logs/{name}/records`  

Upgrade: styx/0  
Connection: Upgrade  

### Params 

| Name             	| In     	| Description                                                                                         	| Default 	|
|------------------	|--------	|-----------------------------------------------------------------------------------------------------	|---------	|
| `name`           	| path   	| Log name.                                                                                           	|         	|
| `X-Styx-Timeout` 	| header 	| The maximum amount of seconds the peer will keep the connection opened whithout receiving messages. 	|         	|

### Response 

```
Status: 101 Switching protocol
```

### Code samples

**Go** (_Requires [TODO Styx client](), [TODO Styx recio](), [TODO Styx api]()  packages._)

```golang
c := client.NewClient("http://localhost:8000")

bufferSize := 1 << 20 // 1 Mb
timeout := 6

params := api.ReadRecordsTCPParams{
	Whence:   log.SeekOrigin,
	Position: 0,
}

reader, err := c.ReadRecordsTCP("test", params, recio.ModeAuto, bufferSize, timeout)
if err != nil {
	logger.Fatal(err)
}

for {
	_, err := reader.Read(&record)
	if err == io.EOF {
		break
	}

	if err != nil {
		logger.Fatal(err)
	}
}
```
