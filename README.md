# Factual Docs

[https://factualdocs.com](https://factualdocs.com)

This web app is made using Golang, HTML, CSS, JavaScript, PostgreSQL and Redis. It is basically a documentary library that automatically fetches and posts videos (documentaries) from predetermined sources (YouTube playlists) therefore it heavily utilizes the [YouTube API](https://developers.google.com/youtube/v3/docs).

It validates the videos against multiple criteria such as:

- not already posted
- 30+ minutes in length
- must be public
- not age restricted
- not region restricted
- embeddable
- must have English audio, title and description, and
- it is not an ongoing or scheduled broadcast

Once the video satisfies all of this criteria it is validated and whitelisted to be automatically posted.

Via a background process a function is periodically called which goes through the playlists (video sources) in the database and checks if there are new videos by using the YouTube API and automatically posts the videos if any. The app is autonomous in that regard. The admin can also manually post videos and of course add new video sources (playlists).

Users can login via Google and Facebook. The app doesn't store passwords so naturally it makes use of the [Google's OAuth 2.0](https://developers.google.com/identity/protocols/oauth2) and [Facebook's Login Flow](https://developers.facebook.com/docs/facebook-login/guides/advanced/manual-flow).


## Run the app locally

Set `DEBUG` to `true` in the `.env` file, as well as populate the entire `.env` file, according to the `example.env`.

Put this alias in your `~/.bash_aliases` file and run `build` whenever you want to build and run the app.
``` bash
alias build='docker compose pull && docker compose up --build --detach'
```

The app will run with [air](https://github.com/air-verse/air) which will provide live reloading. Sometimes though during a long running coding session you might want to rebuild this because `air` can stop live reloading. Access the app on `https://localhost:port` where `port` is the port defined in `PORT` in the `.env` file.

To see the app logs in real time use this:
``` bash
docker compose logs -f app
```

Put this alias in your `~/.bash_aliases` file too and run `down` to bring down the system.
``` bash
alias down='docker compose down --remove-orphans && docker system prune --force'
```

**NOTE**: `worker` and `backup` binaries are being recompiled during each live reload too, so code changes instantly propagate there too.

### Run the worker
``` bash
docker compose exec worker /src/bin/worker
```

### Run the backup
``` bash
docker compose exec backup /src/bin/backup
```

### Access redis
``` bash
docker compose exec -it redis redis-cli
```

### Access the database

`-U` is the user, and `-d` is the database.

``` bash
docker compose exec -it postgres psql -U xxx -d xxx
```


## Run the app in production

No really a difference, except the `app`, `worker` or the `backup` will be built and run by the `Dockerfile` so you need the `TARGET` environment variable in production to specify which one you want to run, `app`, `worker` or `backup`. That is the host needs to be able to pass this `TARGET` variable as a build argument.


## Dump/Restore DB data

Run `export HISTCONTROL=ignorespace` so you're able to hide bash commands from the history if they start with empty space. This is probably already set on your system in the `~/.bashrc` file, but to be sure run it, there's no harm.

Export the database URLs you'll need. Leave an empty space before the export commands. From now on you can use these variables in dump and restoee commands.
``` bash
 export PROD_DB_URL=postgresql://user:password@host:port/dbname
 export LOCAL_DB_URL=postgresql://user:password@localhost:5432/dbname
```

The flag `-a` determines if you want to dump or restore the data with or without the schema. Include `-a` if you want the DATA ONLY.

Dump from production:
``` bash
docker run --rm -e PROD_DB_URL postgres:16.3 pg_dump -F t -d $PROD_DB_URL > db.prod.dump
```

Copy the dump file into the local running container:
``` bash
docker cp ./db.prod.dump postgres:/tmp/db.prod.dump
```

Restore to the local running container:
``` bash
 docker compose exec -e LOCAL_DB_URL postgres pg_restore -F t -d $DATABASE_URL /tmp/db.prod.dump
```

Also, for example, dump from the local running container:
``` bash
docker compose exec -e LOCAL_DB_URL postgres pg_dump -F t -d $DATABASE_URL > db.local.dump
```

And, restore to production:
``` bash
docker run --rm -e PROD_DB_URL \
-v ./db.local.dump:/db.local.dump postgres:16.3 \
pg_restore -F t -d $PROD_DB_URL /db.local.dump
```

When we need tto access the local running postgres server we use `exec`, and when we want to reach the production we use `run` on our own temporary `postgres:16.3` container. The postgres versions need to match though.


Confirm the data is there. This will land you at `psql` in the docker container from where you can see the tables, query the data, etc.
```
docker compose exec -it postgres psql -U xxx -d xxx
```


## Database Golang migration commands

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

## License

[![License: GNU GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg?label=License)](/LICENSE "License: GNU GPLv3")