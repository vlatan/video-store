* User actions

// edit: update content
UPDATE posts SET title = $1, description = $2, updated_at = NOW() WHERE id = $3;

// delete: remove everything
DELETE FROM posts WHERE id = $1 AND user_id = $2;


* Do not parse the content.html for every template
* Modify Cached function to accept a flag whether to return cached or uncached results
  Rename it accordingly
  
* Setup updated_at triggers for the tables
* Delete user account

* Serve Favicons from root
* Error pages
* Protected routes in middleware

* Templates are not minified?
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route
* Error pages