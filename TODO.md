
* Check if you need the auth service if you only use get user from context
  And move get user from context to utils
  Also see if you can move GetUserFromSession to utils

* Pages
  Add Edit button on page
  Slugify

* Sitemaps
  Separate xml templates for pages, sources, categories, misc
  Figure out a way to handle the main index xml page

* Cron

* API calls need retries
* Write custom errors middlware
* Trim services of what related services they don't use
* Add new category form

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Maybe minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route