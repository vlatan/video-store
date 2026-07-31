* Build the frontend and backend review system.

  In prod DO **NOT FORGET** first to:
   - Run the DB migration
   - Iinclude new env var and 
   - Rebuild the setup

  If a review is added we just prepend the review in the review list.
  If a review is updated scroll to the review and highlight it with fade out.
  If a review is deleted we remove it from the list if it's on the first page.
  
  Add number of reviews to html.
  
  Change ids to classes.
  Remove the css vars in rating css.

  Move h4 in the header of the review?
  Use schema on the reviews.

  The user should be able to delete a review or a rating.
  Include delete button on the modals and add edit button on the review card.
  On deletion inline the deletion exalplanation and confirmation in the modals with swapping the modal footer.

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
