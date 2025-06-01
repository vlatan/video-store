* Try to recreate the homepage
* User login
  * Implement flash messages
    You basically need to implement a session just for them and add them to template data struct
    But if there's any caching on the pages you'll use that ability
	* On login add/update user in database and store avatar URL on disk
  * Implement last seen, think about it when to write to DB
  * Figure out a way to retry fetshing avatar if use has default avatar
  * Create analytics ID, base64

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Error pages