# query

Query is a small tool for querying the database and output the result on the screen 

``` sh
go run . -server oracle://user:pass@server/service_name "SELECT *  FROM v$instance"
```
```
INSTANCE_NUMBER          : 1
INSTANCE_NAME            : XE
HOST_NAME                : be7d106dd927
VERSION                  : 11.2.0.2.0
STARTUP_TIME             : 2020-11-11 18:30:05 +0000 UTC
STATUS                   : OPEN
PARALLEL                 : NO
THREAD#                  : 1
ARCHIVER                 : STOPPED
LOG_SWITCH_WAIT          : <nil>
LOGINS                   : ALLOWED
SHUTDOWN_PENDING         : NO
DATABASE_STATUS          : ACTIVE
INSTANCE_ROLE            : PRIMARY_INSTANCE
ACTIVE_STATE             : NORMAL
BLOCKED                  : NO
EDITION                  : XE
```
