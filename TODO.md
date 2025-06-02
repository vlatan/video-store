* Try to recreate the homepage
* User login
  * Redirect back the user if user clicks cancel in google/fb login flow
  * Send failed login/logout flashes
	* On login store avatar URL on disk
  * Implement last seen, think about it when to write to DB
  * Figure out a way to periodically retry downloading avatar if the user has default avatar

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Error pages