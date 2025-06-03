* Try to recreate the homepage
* User login
	* On login store avatar URL on disk
  * Store user key in redis for the avatar, retry download periodically if serving default avatar
  * Serve avatars from file server on the static handler
  * Implement last seen on user, think about it when to write to DB

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages