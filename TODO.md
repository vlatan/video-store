* Add `original_title` column in the `post` table.
  That will entail changing almost all queries to SELECT/UPSERT the original title too.
  As well as adding mechanism that if the there's no original title to default to the video title.
  That can be done with a method on the Post struct to fetch the official title.
  Also keep using the video title as video title.

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
