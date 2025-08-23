* Store avatar on R2
  Practically similar implementation of GetAvatar but disk is replaced with R2
  And the checking of file existance on disk is replcaed with HEAD request to R2
  The only difference I upload the default avatar too to R2 where needed

* Check the memory leaks
* Check for nil dereference in templates
* Add slug input to page
* Add form for new category as well as delete, edit category routes
* Delete source - should cascade and delete all the videos?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
