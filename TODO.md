* Make CF cache pages for logged-out users
  Send headers
  Also create a cache rule to skip caching requests with session cookies
  Maybe you'll need to NOT clear the flash cookie immediately

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
