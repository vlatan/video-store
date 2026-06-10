* OAuth verification homepage
* What to do with no candidates videos due to prohibited content?
* Maybe use another gemini API call for the credits

* User reviews and ratings
* Search bar on small screens accross entire screen
* Group videos by entity, create taxonomies
* Create "Best of" landing pages for these clusters
* Internal linking

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
