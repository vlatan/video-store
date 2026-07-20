* Build the frontend and backend review system.
  SMake quick rating feature a form.
  Include big star in the review form.
  Change ids to classes.

  User should be able to post just one review per post.
  Schema reflects that, there's UNIQUE(user_id, post_id).
  So maybe preserve/populate the user review in the dialog and if they want to edit, update in DB.
  Consequently the user should be able to delete a review or a rating for that matter.
  Maybe add delete button in the dialog next to the submit.
  That is doable on the review dialog but not so much on the rating dialog.

  Show the rating on the review if any or make the rating option available on the review.
  Decide whether to show more reviews with "load more", infinite scroll or with pagination.

* Make the checkmarks on the like/save green or yellow
* Add close button to login menu
* Make search bar on small screens accross entire screen

* Maybe use another gemini API call for the credits
* Group videos by entity, create taxonomies
* Create "Best of" landing pages for these clusters
* Internal linking

* Eventually remove tags and description from search vector
* Write tests

* Minify CSS and JS files during development.
  Calculate just etags on compile or on the fly in the route

* Refactor JS in functions and classes
  Use TypeScript
