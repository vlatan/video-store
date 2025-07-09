* Sources
  * Single source posts handler (should it be at the sources handlers and use posts repo?)
    The same with the single category posts handler?
    The same with the search?

  * Since the source (playlist) is showing the channel thumbnail two playlists
    from the same channel will be shown with the same thumbnail. Maybe add or use the
    playlist title if it's not the entire channel uploads playlist.

* Pages
* Admin (get users with pagination)
* Sitemap
* Cron
* User watch later

* Write custom errors middlware

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route