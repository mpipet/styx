Getting started
---------------

### Quick Install

```bash

```

### Run Styx

```bash

```

### Create a log
```bash
$ styx logs create myLog
name:                   myLog
status:                 ok
record_count:           0
file_size:              0
start_position:         0
end_position:           0
```

### Write records
```bash
$ styx logs write myLog
>my first record
>my second record
```

### Read records
```bash
$ styx logs read myLog
my first record
my second record
```
