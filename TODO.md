* Try to recreate the homepage
* User login
	* On login add/update user in database and store avatar URL on disk
  * Create analytics ID
  * Sort out the login pop up
  * Pass the state (redirect url) to flow and retrieve at the end
    https://claude.ai/chat/8454a3b9-7cce-444b-a146-e34ec937baac
    
    Needs current URL in the auth and logout links in templates
  * Pass FlashMessages to redirect

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Error pages