* Get random videos as related if none
* Make redis cache function to return T

* Setup updated_at triggers for the tables
  https://claude.ai/chat/49e95de4-f9f6-4d04-a5d5-dab613c0ae93

* Write custom errors middlware,
  Detect html requests and return custom HTML errors
  https://claude.ai/chat/80a403c0-994d-4377-8c47-b5087a6e6af1
  
* User avatar URL exists in redis but not locally
  Need to solve this discrepancy

##############################################################

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
* Add slug input to page
* Add new category form as well as delete, edit routes
* Check for nil values in templates