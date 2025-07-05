* On auth middlwares get current user will add user in context only if auth or admin,
  otherwise will serve forbidden.

  But then we are getting the current user again in the handler,
  and if not in context we're going to get it from session.
  Think about this.

* Solve redirect destination on logout on a forbidden route
  Track forbidden routes and if redirect destination one of them go home.
  For example you are on a new video route and you logout, the redirect
  will be on that route but you'll get foribidden.

* Post new video
* Sources
* Pages
* Admin (get users with pagination)
* Sitemap
* Cron

* Write custom errors middlware

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route