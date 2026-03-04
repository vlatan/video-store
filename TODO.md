* Run new migration file
* Addapt the search query
* Make gaps between `yt-dlp` calls 10 minutes, basically between videos
* Apply the DB migration to production. Bind the remote database to localhost.
  `ssh -L 5432:localhost:5432 user@your-vps`

* Delete sources and posts without credits - mainly video essays
  Should post cascade delete if source is deleted?

* Add markup schema for the credits
  The YT video title should go to the VideoObject
  Do not repeat persons names/bios in different roles

* Group videos by entity, create taxonomies
* Create "Best of" landing pages for these clusters
* Internal linking

* Uncomment and remove code in:
  - `internal/integrations/gemini/gemini.go`
  - `internal/worker/worker.go`
  - `internal/utils/utils.go`

* Write tests
* Add slug input to page
* Add form for new category as well as delete, edit category routes
* Delete source - should cascade and delete all the videos?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
  Use TypeScript
