
* Sitemap
  Minify the base and content as starting templates and proceed to minify the rest.
  Solve +html in RenderHTML, maybe separate function for RenderXML, RenderXSL
  Maybe use the names including the extension.

* Pages
* Admin (get users with pagination)
* Cron
* User watch later

* API calls need retries

* Write custom errors middlware

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route