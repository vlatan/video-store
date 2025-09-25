* For PublicCache middleware the user is NOT loaded
  because in the LoadUser middlware we exclude if URL is sitemap,
  so the check for IsCurrentUserAdmin in the PublicCache will always fail
  so the Cache-Control header will always be there.

* We also load the data in context for all the static files.
  That's not necessary.

* Rename the NeedsSession func to something like IsFilePath.
  We need to check for extension and possibly allow sitemap files (.xml, .xsl)

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
