# Video Store

[https://factualdocs.com](https://factualdocs.com)

This web app is made using Golang, HTML, CSS, JavaScript, PostgreSQL and Redis. It is basically a video library that automatically fetches and posts videos from predetermined sources (YouTube playlists) therefore it heavily utilizes the [YouTube API](https://developers.google.com/youtube/v3/docs).

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

The quickest way to mimic locally-trusted HTTPS access is to install a CA (Certificate Authority) and generate certificates signed with that authority. There are several tools that can do this very easy but we'll use `mkcert`.

Install `mkcert` and create and install a local CA in the system root store.

```
sudo apt install libnss3-tools
go install filippo.io/mkcert@latest
mkcert -install
```

Generate a private key and certificate for the same **DOMAIN** defined in your `.env` file. The snippet below generates a certificate for `dev.domain.com` as well as for `"*.dev.domain.com"`.

```
DOMAIN=dev.domain.com && \
DIR=certs && \
mkdir -p $DIR && \
mkcert -key-file $DIR/local.key \
-cert-file $DIR/local.crt \
$DOMAIN "*.${DOMAIN}"
```

Also edit your local `/etc/hosts` and map `127.0.0.1` to `dev.domain.com` and `dash.dev.domain.com`.

```
# /etc/hosts
127.0.0.1       dev.domain.com
127.0.0.1       dash.dev.domain.com
```

In this example the app can be accessed at `https://dev.domain.com` and the traefik dashboard at `https://dash.dev.domain.com`.

The secret keys (`CSRF_KEY`, `AUTH_KEY`, `ENCRYPTION_KEY`) and the `GEMINI_PROMPT` need to be `base64` encoded.

For the secret keys you can use the following recommended snippet from `gorilla/sessions` to generate different keys and encode them to `base64`:
``` golang
key := securecookie.GenerateRandomKey(32)
log.Println(base64.StdEncoding.EncodeToString(key))
```

For the Gemini prompt here's the exact `JSON` structure with example text which you'll need to `base64` encode by using `base64.StdEncoding.EncodeToString(prompt)` as well. The placeholders need to be in the exact same format as shown below because the code will look for them to replace them.
``` json
[
    {
        "text": "Summarize the video."
    },
    {
        "text": "Select one category from these categories: {{ CATEGORIES }}."
    },
    {
        "url": "https://www.youtube.com/watch?v={{ VIDEO_ID }}",
        "mime_type": "video/mp4"
    }
]
```


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

### Run the worker
``` bash
docker compose run --rm --build worker
```

### Run the backup
``` bash
docker compose run --rm --build backup
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


## Useful information for identifying memory leaks

While logged in as admin visit the `/debug/heap` endpoint and the file will download. Rename it to `heap1`.  
Stress the app locally by using:
``` bash
for i in {1..5000}; do curl -s http://localhost:5000/ > /dev/null; done &
for i in {1..5000}; do curl -s http://localhost:5000/ > /dev/null; done &
for i in {1..5000}; do curl -s http://localhost:5000/ > /dev/null; done &
wait
```
Visit the `/debug/heap` endpoint again and download another heap profile file. Rename it to `heap2`.  
Run a profile which will compare the base (`heap1`) and the next profile (`heap2`):
``` bash
go tool pprof -base heap1 heap2
```

Once inside the `pprof` CLI run `top`, or `top20` to see if there's memory increase.


## Run tests

Produce coverage report and heat map.

``` bash
go test -race -coverprofile=coverage.out ./... && 
go tool cover -html=coverage.out
```

Target specific package.
``` bash
go test -race -coverprofile=coverage.out ./internal/integrations/gemini && 
go tool cover -html=coverage.out
```



## Dump/Restore DB data

Run `export HISTCONTROL=ignorespace` so you're able to hide bash commands from the history if they start with empty space. This is probably already set on your system in the `~/.bashrc` file, but to be sure run it, there's no harm.

Export the database URL vars you'll need. Leave an empty space before the export commands. From now on you can use these variables in dump and restore commands.
``` bash
 export PROD_DB_URL=postgresql://user:password@host:port/dbname
```

The flag `-a` determines if you want to dump or restore the data with or without the schema. Include `-a` if you want the **DATA ONLY**.  
These are the available dump formats.

``` bash
-Fc = custom format (compressed, flexible)
-Ft = tar format
-Fp = plain SQL
-Fd = directory format
```

Dump from a remote database using URL. On the fly we're creating, running - and when done removing - our own temporary `postgres:16.3` container to match the remote postgres version.
``` bash
docker run --rm -e PROD_DB_URL postgres:16.3 pg_dump -Fc -d $PROD_DB_URL > db.dump
```

Dump from a running postgres container. Note, we're using `exec` to execute a command within the container.
``` bash
docker compose exec -T <service_name> pg_dump -U <user> -Fc <db_name> > db.dump
```

Restore the database if the dump is plain SQL.
``` bash
docker compose exec -T <service_name> psql -U <user> -d <user> < db.dump
```

Restore the database if the dump is NOT plain SQL.
``` bash
docker compose exec -T <service_name> pg_restore -U <user> -d <db_name> --clean --no-owner < db.dump
```


## Database Golang migration commands

Here's a little [PostgreSQL golang-migrate tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md).

Install the `golang-migrate` library.
``` bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Export the database URL in a variable. The empty space at the beginning is important so the command is not saved in the bash history. Change the values in the URL string accordingly.
``` bash
 export DATABASE_URL='postgres://postgres:password@localhost:5432/example?sslmode=disable'
```

Check the current version.
``` bash
migrate -path migrations -database $DATABASE_URL version
```

Create a specific migration file.
``` bash
migrate create -ext sql -dir migrations -seq file_name
```

Migrate up/forward. Supply how many steps you want to go up.
``` bash
migrate -path migrations -database $DATABASE_URL up <steps>
```

Migrate down/backward. Supply how many steps you want to go down.
``` bash
migrate -path migrations -database $DATABASE_URL down <steps>
```

Force a version. Useful for rollback from a dirty version to the previous version and trying again, or if you want to force a current dirty version you are sure it went okay.
``` bash
migrate -path migrations -database $DATABASE_URL force <version_number>
```

## License

[![License: GNU GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg?label=License)](/LICENSE "License: GNU GPLv3")