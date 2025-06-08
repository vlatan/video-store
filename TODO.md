* Validate video ID
* Make ready the single post template
* Finish the single post handler
* Refactor the db post helper functions
* Possibly change how the templates are parsed.
  First parse the layout.
  The layout needs to be called in each partial template
  Then parse all the partials.
  
* Delete user account

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages