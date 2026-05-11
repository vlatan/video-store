* First try updating JUST the summary and title
  Change CI to use the feature branch
  Push code changes
  Do not process very large videos
  Make gaps between `yt-dlp` calls 10 minutes, basically between videos
  Test this setup for a prolonged period of time in prod
  If it's not working revert to main branch in CI, abandon this feature.

* If everything goes well

* Add title in Credits struct and use this as na actual Post title.
  Probably add VideoTitle in the Post struct and the DB
  to use it as a title in the VideoObject markup.
  Run the DB migration locally

  Adapt thet single post query (SELECT AND UPSERT)
  Addapt the search query

* Sanitize the incoming parts.
  Gemini returned no candidates, reason=PROHIBITED_CONTENT
  In this case the audio or images might be overly explicit.

* Add markup schema in the post template for the credits.
  Do not repeat persons names/bios in different roles.
  Design the sections better.

* Run the DB migration to production.
  Bind the remote database to localhost:
  `ssh -L 5432:localhost:5432 user@your-vps`
  Update the README.md for this process.

* Update README.md in general, remova stale info.

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
