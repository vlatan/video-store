# Project factual-docs

One Paragraph of project description goes here

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```


## Dump db data from the remote host and restore it locally to docker

Run `export HISTCONTROL=ignorespace` so you're able to hide bash commands from the history if they start with empty space. This is probably already set on your system in the `~/.bashrc` file, but to be sure run it, there's no harm. Run the following commands starting with empty space.

First, dump the database (data only `-a`) from the remote host in a tar format (`-F t`). The postgres version (postgres:16.3) needs to match the remote version.
```
 docker run --rm -e PGPASSWORD=xxx \
postgres:16.3 pg_dump -U xxx -h xxx -p xxx -a -F t xxx > db.dump
```

Copy the dump into the running local postgres container, which also has to match the version.
```
 docker cp ./db.dump postgres:/tmp/db.dump
```

Execute the restore. Restory only the data (`-a`).

```
 docker compose exec -e PGPASSWORD="xxx \
postgres pg_restore -U xxx -h localhost -p 5432 -d xxx -a -F t /tmp/db.dump
```

Confirm the data is there. This will land you at `psql` in the docker container from where you can see the tables, query the data, etc.
```
docker compose exec -it postgres psql -U xxx -d xxx
```


## Migration commands

Here's a little [PostgreSQL golang-migrate tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md).

Install the `golang-migrate` library.
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Export the database URL in a variable. The empty space at the beginning is important so the command is not saved in the bash history. Change the values in the URL string accordingly.
```
 export DATABASE_URL='postgres://postgres:password@localhost:5432/example?sslmode=disable'
```

Check the current version.
```
migrate -path migrations -database $DATABASE_URL version
```

Create a specific migration file.
```
migrate create -ext sql -dir migrations -seq file_name
```

Migrate up/forward. Supply how many steps you want to go up.
```
migrate -path migrations -database $DATABASE_URL up <steps>
```

Migrate down/backward. Supply how many steps you want to go down.
```
migrate -path migrations -database $DATABASE_URL down <steps>
```

Force a version. Useful for rollback from a dirty version to the previous version and trying again, or if you want to force a current dirty version you are sure it went okay.
```
migrate -path migrations -database $DATABASE_URL force <version_number>
```