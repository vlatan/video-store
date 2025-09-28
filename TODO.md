* Testing packages in parallel will probably
  create race conditions, when resetting the singleton db service

* Fix GetProjectRoot 
  It will work only if the file from where it's run
  is two dirs down from the project root

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
