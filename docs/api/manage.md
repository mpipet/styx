Manage logs
-----------

### Create log

Create a new log

```bash
curl -XPOST 'http://localhost:8000/logs' -d name=myLog
```

Response

```json
{
  "name": "myLog",
  "status": "ok",
  "record_count": 0,
  "file_size": 0,
  "start_position": 0,
  "end_position": 0
}
```

List of all available form parameters:

| Param               | Description                                                         | Default      |
|---------------------|---------------------------------------------------------------------|--------------|
| `name`  _required_  | The log name.                                                       |              |
| `max_record_size`   | Max record size.                                                    | `1048576`    |
| `index_after_size`  | Allow creating an index entry every index_after_size bytes written. | `1048576`    |
| `segment_max_count` | Max number of records in a segment.                                 | `-1`         |
| `segment_max_size`  | Max size of a segment in bytes.                                     | `1073741824` |
| `segment_max_age`   | Max age of a segment in seconds.                                    | `-1`         |
| `log_max_count`     | Max number of records in a log.                                     | `-1`         |
| `log_max_size`      | Max size of a log in bytes.                                         | `-1`         |
| `log_max_age`       | Max age of a log in seconds.                                        | `-1`         |

### List logs

Retrieves the details of all Styx logs

```bash
curl -XGET 'http://localhost:8000/logs'
```

Response

```json
[
  {
    "name": "myLog",
    "status": "ok",
    "record_count": 1345,
    "file_size": 1845,
    "start_position": 500,
    "end_position": 845
  },
  {
    "name": "myOtherLog",
    "status": "ok",
    "record_count": 542,
    "file_size": 730,
    "start_position": 0,
    "end_position": 542
  },
]
```

### Get log by name

Retrieves the details of a log

```bash
curl -XGET 'http://localhost:8000/logs/myLog'
```

Response

```json
{
  "name": "myLog",
  "status": "ok",
  "record_count": 1345,
  "file_size": 1845,
  "start_position": 500,
  "end_position": 845
}
```

### Delete log

Permanently delete a log and its data

```bash
curl -XDELETE 'http://localhost:8000/logs/myLog'
```

### Backup log

Download a backup of the log

```bash
curl -XGET 'http://localhost:8000/logs/myLog/backup' -o myLogBackup.tar.gz
```

### Restore log

Imports a previously backed up log archive

```bash
curl -XPOST 'http://localhost:8000/logs/restore?name=myRestoredLog' --data-binary '@myLogBackup.tar.gz'  
```
