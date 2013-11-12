# revisioneer

create deployment timelines to communicate changes easier with your clients.

### TODO

- add tests
- add multi tenancy support /w api tokens as authentication (project 1 - * deployment)
- add support for deployment changesets (summary of git commit message headlines) (deployment 1 - * changes)
- add database configurations via yml/ json (read at start up)

### Examples

**Create a new revision**
curl -X POST "http://127.0.0.1:8080/revisions" -d '{ "sha": "asdasd" }'

**Read all revisions**
curl "http://localhost:8080/revisions"