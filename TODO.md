* Apply the migration in production
* Add new ENV vars
* Push the code change

* Drop the old colums (create migration)
* Apply that migration to production too

* Decide on whether to make the email unique
  Drop users with no email
  On login check for email and overwrite the authID and provider
  If no email create new account

* Check for nil dereference in templates
* Add slug input to page
* Add form for new category as well as delete, edit category routes
* Delete source - should cascade and delete all the videos?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
