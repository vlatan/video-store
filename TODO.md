
* Order promot so the URL part comes last
* Include the update marker in manualy posted videos too

* Uncomment and remove code in:
  - `internal/worker/utils.go`
  - `internal/worker/worker.go`
  - `internal/utils/utils.go`

* Generalize the app.
  Eliminate the documentary usage.

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
