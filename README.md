# Factual Docs

[https://factualdocs.com](https://factualdocs.com)

This web app is made using Python (Flask), HTML, CSS, JavaScript, PostgreSQL and Redis. It is basically a documentary library that automatically fetches and posts videos (documentaries) from predetermined sources (YouTube playlists) therefore it heavily utilizes the [YouTube API](https://developers.google.com/youtube/v3/docs).

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

To run the worker use this. Worker binary is being recompiled during each live reload too, so code changes instantly propagate there too.
``` bash
docker compose run --rm worker /src/bin/worker
```

To access redis use this:
``` bash
docker compose exec -it redis redis-cli
```

To access the database use this, where `-U` is the user, and `-d` is the database.
``` bash
docker compose exec -it postgres psql -U xxx -d xxx
```

## Dump db data from the remote host and restore it locally to docker

Run `export HISTCONTROL=ignorespace` so you're able to hide bash commands from the history if they start with empty space. This is probably already set on your system in the `~/.bashrc` file, but to be sure run it, there's no harm.

First, dump the database (data only `-a`) from the remote host in a tar format (`-F t`). The postgres version (postgres:16.3) needs to match the remote version.
```
 docker run --rm -e PGPASSWORD=xxx \
postgres:16.3 pg_dump -U xxx -h xxx -p xxx -a -F t xxx > db.dump
```

Copy the dump into the running local postgres container, which also has to match the version.
```
 docker cp ./db.dump postgres:/tmp/db.dump
```

Execute the restore. Restore only the data (`-a`).

```
 docker compose exec -e PGPASSWORD=xxx \
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

## License

[![License: GNU GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg?label=License)](/LICENSE "License: GNU GPLv3")