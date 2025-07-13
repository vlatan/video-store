
* Users in dashboard
  Get local avatars if ANY concurrently
  Format dates on users dash
  Adjust index to start at 1
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