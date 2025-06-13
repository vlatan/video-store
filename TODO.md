* Remove "&", "|", "(", ")" from search query because it will fail
  ERROR: no operand in tsquery:

* Do not parse the content.html for every template
* Modify Cached function to accept a flag whether to return cached or uncached results
  Rename it accordingly
* Maybe add similarity pg_trgm boost on the the short description

* Delete user account

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages