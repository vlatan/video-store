
* Users in dashboard

  Get local avatars if ANY concurrently
  There are users with empty analytics IDs and the first user with empty analytics ID
  is saved with "avatar:" redis key and then all the users that
  don't have analytics ID retrieve this same avatar as theirs.

  Provide paginated view (scroll or pagination)

* Pages

* Sitemaps
  Separate xml templates for pages, sources, categories, misc
  Figure out a way to handle the main index xml page

* Cron

* API calls need retries
* Write custom errors middlware
* Trim services of what related services they don't use

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Maybe minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route