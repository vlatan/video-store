* Run the DB migration locally
* Addapt the search query

* Add markup schema in the post template for the credits.
  The YT video title should go to the VideoObject.
  Do not repeat persons names/bios in different roles.
  Design the sections better.

* Do not process very large videos
* Make gaps between `yt-dlp` calls 10 minutes, basically between videos

* Run the DB migration to production.
  Bind the remote database to localhost:
  `ssh -L 5432:localhost:5432 user@your-vps`
  Update the README.md for this process.

* Push code changes

* Delete sources and posts without credits - mainly video essays
  Should post cascade delete if source is deleted?
  If so another DB migration is required.

* Group videos by entity, create taxonomies
* Create "Best of" landing pages for these clusters
* Internal linking
* Rating
* User reviews

* Uncomment and remove code in:
  - `internal/integrations/gemini/gemini.go`
  - `internal/worker/worker.go`
  - `internal/utils/utils.go`

* Eventually remove tags and description from search vector

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
