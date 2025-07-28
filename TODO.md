
* Does not set Vary header
  w.Header().Set("Vary", "Accept-Encoding")

* Static files GZIP and bypass
  https://claude.ai/chat/d9ead877-6400-4213-a927-33cc5ded9fdd 
  https://www.lemoda.net/go/gzip-handler/index.html
  https://github.com/klauspost/compress

* Do not set session and flash cookie on static files

* Set csfr cookie only for auth or admin users
  I should probably not create the csrf.Protect middlware at all in this case
  because that produces the cookie.
  Check if I should do this at all?
  https://claude.ai/chat/6757dc08-0008-4ea6-a6b2-91ea949e9aa1

* The cloudflare cdn header, once we send it it's done for everyone?
* Get from the backup a db and see how many orphans there were
* Delete the python convert script and everything with it

* Check for nil dereference in templates

* Maybe the worker to backup the database?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
* Add slug input to page
* Add new category form as well as delete, edit routes
* Check for nil values in templates