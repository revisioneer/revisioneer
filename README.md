# revisioneer

create deployment timelines to communicate changes to your clients.
revisioneer is a backend written in Go which sole purpose is to store your
deployments & changeset messages you want to communicate.

You can retrieve those informations at any time, but how you display them
is entirely up to you.

The service is provisioned using ansible. So take a look at [revisions-provisioning](https://github.com/nicolai86/revisions-provisioning) if you want to run your own.

## Tests

To run the testsuite you need to have a PostgreSQL server running & deployed. Revisioneer uses [sqitch][1] for schema management. Thus you need to run

``` bash
createdb revisioneer_test
sqitch -d revisioneer_test deploy
```

```
REV_DSN="user=$(whoami) dbname=revisioneer_test sslmode=disable" go test ./...
```

## Executing

``` bash
createdb revisioneer
sqitch deploy

gom install
go build
REV_DSN="user=$(whoami) dbname=revisioneer sslmode=disable" ./revisioneer
```

### API Examples

#### Create a project

    curl -X POST "http://127.0.0.1:8080/projects" -d '{ "name": "test" }'
    # => 200 OK
    {
       "name": "test",
       "api_token": "q+fehEVx5Kxast2DdUUnKaQpNiZ4GTsmmaYerNwDXDE=",
       "created_at": "2013-11-14T22:48:54.431707172+01:00"
    }

Make sure to keep the `api_token` around. There is currently no way to retrieve it.

#### Create a new deployment information

    curl -X POST "http://127.0.0.1:8080/deployments" \
      -d '{ "sha": "61722b0020", "messages": ["* added support for messages"], "new_commit_counter": 1 }' \
      -H "API-TOKEN: q+fehEVx5Kxast2DdUUnKaQpNiZ4GTsmmaYerNwDXDE="
    # => 200 OK

#### Verify a deployment

    curl -X POST "http://127.0.0.1:8080/deployments/61722b0020/verify" \
      -H "API-TOKEN: q+fehEVx5Kxast2DdUUnKaQpNiZ4GTsmmaYerNwDXDE="
    # => 200 OK

#### Read all deployments

    curl "http://localhost:8080/deployments" \
      -H "API-TOKEN: q+fehEVx5Kxast2DdUUnKaQpNiZ4GTsmmaYerNwDXDE="
    # => 200 OK
    [
        {
            "sha": "61722b0020",
            "deployed_at": "2013-11-14T22:52:40.746848+01:00",
            "messages": [
                "* added support for messages"
            ]
        }
    ]

Returns only the most recent 20 deployments. You can adjust this using `page` and `limit` parameters.

[1]:https://github.com/theory/sqitch
