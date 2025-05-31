* Try to recreate the homepage
* User login
	* On login add/update user in database and store avatar URL in redis
  * Pass that struct to templates via middlware request context
  * Pass the state (redirect url) to flow and retrieve at the end
  * Pass FlashMessages to redirect

* Templates are not minified?
* Error pages