Write with Styx protocol
------------------------

Write records using [Styx protocol](/docs/api/styx_protocol.md).

**POST** `/logs/{name}/records`  

Upgrade: styx/0  
Connection: Upgrade  

### Params 

| Name             	| In     	| Description                                                                                         	| Default 	|
|------------------	|--------	|-----------------------------------------------------------------------------------------------------	|---------	|
| `name`           	| path   	| Log name.                                                                                           	|         	|
| `X-Styx-Timeout` 	| header 	| The maximum amount of seconds the peer will keep the connection opened whithout receiving messages. 	|         	|


### Code samples

**Go** (_Requires [TODO Styx client](), [TODO Styx recio]()  packages._)

```golang
bufferSize := 1 << 20 // 1Mb
timeout := 10

c := client.NewClient("http://127.0.0.1:8000")
writer, err := c.WriteRecordsTCP("myLog", recio.ModeAuto, bufferSize, timeout)
if err != nil {
	logger.Fatal(err)
}

record := log.Record("my record content")

for i := 0; i < 10; i++ {

  _, err = writer.Write(&record)
  if err != nil {
  	logger.Fatal(err)
  }
}

err = writer.Flush()
if err != nil {
	logger.Fatal(err)
}
```
