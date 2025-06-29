* User avatar saved to redis not locally
  Need to solve this discrepancy

* Trailing slash 301 redirect on login/logout
  SanitizeRelativePath "cleans" the path removing the trailing slash.
  Which is good for sanitizing paths on the static handler,
  but not good for getting user redirect.

* Construct WWW to non-WWW redirect middleware

* Construct absolute canonical URL, not relative

* ads.txt route

* Serve Favicons from root

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route