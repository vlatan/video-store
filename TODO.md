* Try to recreate the homepage
* User login
	* On login store avatar URL on disk
  * Implement last seen, think about it when to write to DB
  * Figure out a way to periodically retry downloading avatar if the user has default avatar

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages