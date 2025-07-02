* Domain restructure:
  https://claude.ai/chat/fa47abbb-09f0-401e-b118-c24e72d2fb01

* Consider moving files out of shared because it has a handler

* User avatar saved to redis not locally
  Need to solve this discrepancy

* Setup updated_at triggers for the tables
* Minify CSS and JS files before deployment and save them before embedding
  Calculate just etags on init or on the fly in the route