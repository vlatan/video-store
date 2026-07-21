* Build the frontend and backend review system.
  Include big star in the review form.
  Change ids to classes.

  User should be able to post just one review per post.
  Schema reflects that, there's UNIQUE(user_id, post_id) so upsert the review/rating.
  Update the schema so the review can't exist without a rating.
  Cascade delete review if the rating was deleted.
  So when deleting we need to delete just the rating and if there's a review that will be deleted too.

  Preserve/populate the user review/rating in the dialog and if they want to edit, update in DB.
  Consequently the user should be able to delete a review or a rating for that matter.
  Include delete button on the modals and add edit button on the review card.
  If the user has a review the "quick rate" button opens the review modal.
  On deletion inline the deletion exalplanation and confirmation in the modals with swapping the modal footer.

  Use "show more reviews" to uncover more reviews.
  Probably use cursor.

  If a review is added we just prepend the review in the review list.
  If a review is updated scroll to the review and highlight it with fade out.
  If a review is deleted we remove it from the list if it't on the first page.

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
