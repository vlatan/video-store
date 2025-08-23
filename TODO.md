* Store avatar on R2

  1. Check Redis "avatar:user:xxxx" 
  2. If exists: serve R2 URL (assume it's there)
  3. If expired/missing:
    - Download from remote source
    - Upload to R2 (custom or default on failure)  
    - Set Redis key with 24hrs, 12hrs, 6hrs expiry
    - Serve R2 URL
  
  If someone deletes the avatar from the bucket the user will
  have a nroken avatar of the redis key expiry time at most. 

* Check the memory leaks
* Check for nil dereference in templates
* Add slug input to page
* Add form for new category as well as delete, edit category routes
* Delete source - should cascade and delete all the videos?

* Use logger
  https://go.dev/blog/slog

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
