* User avatar saved to redis not locally
  Should check if avatar already exists before download

* Construct WWW to non-WWW redirect middleware

* ads.txt route

* Serve Favicons from root

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route