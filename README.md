# revisioneer

create deployment timelines to communicate changes easier with your clients.

## Tests

To run the testsuite you need to have a PostgreSQL server running & deployed. Revisioneer uses [sqitch][1] for schema management. Thus you need to run

``` bash
createdb revisioneer_test
sqitch -d revisioneer_test deploy
```

Then you can use `go` to run the testsuite:

```
REV_DSN="user=$(whoami) dbname=revisioneer_test sslmode=disable" go test
```

## Executing

``` bash
createdb revisioneer
sqitch deploy

REV_DSN="user=$(whoami) dbname=revisioneer sslmode=disable" go run revisioneer.go
```

### TODO

- add support for deployment changesets (summary of git commit message headlines) (deployment 1 - * changes)

### Examples

**Create a project**
curl -X POST "http://127.0.0.1:8080/projects" -d '{ "name": "test", "apiToken": "test" }'

**Create a new revision**
curl -X POST "http://127.0.0.1:8080/revisions" -d '{ "sha": "asdasd" }' -H "API-TOKEN: test"

**Read all revisions**
curl "http://localhost:8080/revisions"

[1]:https://github.com/theory/sqitch