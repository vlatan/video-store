* Fix edit button

* Move avatars out of the web because they are
  ending up in the docker container

* Get random videos as related if none
* Add back the orphaned videos

* Cron
* API calls need retries
  https://github.com/cenkalti/backoff/blob/v5/example_test.go
  Attach orphan post to playlist

* Use logger
  https://go.dev/blog/slog

* Setup updated_at triggers for the tables
  https://claude.ai/chat/49e95de4-f9f6-4d04-a5d5-dab613c0ae93

* Write custom errors middlware,
  Detect html requests and return custom HTML errors
  
* Trim services of what related services they don't use

* User avatar URL exists in redis but not locally
  Need to solve this discrepancy

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
* Add slug input to page
* Add new category form as well as delete, edit routes
* Check for nil values in templates