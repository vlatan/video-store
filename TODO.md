* Create new tables, triggers, functions via the migration file
* Addapt the search query
* Add markup schema if available for the credits
* Make gaps between `yt-dlp` calls 10 minutes, basically between videos
* Apply the migration to production. Bind the remote database to localhost.
  `ssh -L 5432:localhost:5432 user@your-vps`

* Group videos by entity, create taxonomies.
* Create "Best of" landing pages for these clusters.
* Internal linking.

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
