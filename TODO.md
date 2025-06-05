* Try to recreate the homepage
* User login
  * Implement last seen on user
    On login write to session two timestamps
    One when the last DB write was done
    One update on every page load
    On page load compare if these are different dates and if yes only then write to DB
    https://claude.ai/chat/899b9f9c-bdd3-4b04-b261-9973ae02eccd
  * On login distinguish if the user was updated or inserted so to write to updated_at

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages