package posts

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
	"golang.org/x/sync/errgroup"
)

// Handle a single post
func (s *Service) SinglePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get video id from URL path
	videoID := r.PathValue("video")

	// Get template data
	data := s.ui.TmplData(w, r)

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	var (
		post         types.Post
		postReviews  types.Reviews
		userActions  types.Actions
		relatedPosts []types.Post
	)

	g := new(errgroup.Group)

	// Send post fetch in a goroutine
	g.Go(func() error {

		// Don't cache single post for logged in users
		var taskErr error
		if data.CurrentUser.IsAuthenticated() {
			post, taskErr = s.postsRepo.GetSinglePost(r.Context(), videoID)
		} else {
			post, taskErr = rdb.GetCachedData(
				r.Context(),
				s.rdb,
				fmt.Sprintf(postCacheKey, videoID),
				s.config.CacheTimeout,
				func() (types.Post, error) {
					return s.postsRepo.GetSinglePost(r.Context(), videoID)
				},
			)
		}

		return taskErr
	})

	// Get reviews from DB in a goroutine
	g.Go(func() error {

		// Get post reviews, don't cache the reviews for logged in users
		var taskErr error
		if data.CurrentUser.IsAuthenticated() {
			postReviews, taskErr = s.postsRepo.GetPostReviews(r.Context(), videoID, "")
		} else {
			postReviews, taskErr = rdb.GetCachedData(
				r.Context(),
				s.rdb,
				fmt.Sprintf(postReviewsCacheKey, videoID),
				s.config.CacheTimeout,
				func() (types.Reviews, error) {
					return s.postsRepo.GetPostReviews(r.Context(), videoID, "")
				},
			)
		}

		if taskErr != nil {
			return taskErr
		}

		// Get the user avatars
		for i, review := range postReviews.Items {
			localAvatarURL, taskErr := s.avatar.Get(r.Context(), &review.User)
			if taskErr != nil {
				return taskErr
			}
			postReviews.Items[i].User.LocalAvatarURL = localAvatarURL
		}

		// Check if the current user owns a review
		for i, review := range postReviews.Items {
			if data.CurrentUser.IsAuthenticated() && data.CurrentUser.ID == review.User.ID {
				postReviews.Items[i].IsCurrentUser = true
			}
		}

		return nil
	})

	// Wait for the goroutines to finish
	if err := g.Wait(); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(
				r.Context(), "no such post in DB",
				"error", err,
			)
			s.ui.HTMLError(w, r, data, http.StatusNotFound)
			return
		}

		slog.ErrorContext(
			r.Context(), "failed to get a single post",
			"error", err,
		)

		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Get current user actions in a goroutine
	g.Go(func() error {

		// Check whether the current user liked, faved, rated and/or reviewed the post
		if data.CurrentUser.IsAuthenticated() {

			var taskErr error
			userActions, taskErr = s.usersRepo.GetUserActions(
				r.Context(),
				data.CurrentUser.ID,
				post.ID,
			)

			if taskErr != nil {
				return taskErr
			}
		}

		if userActions.Review.Headline != "" && userActions.Review.Content != "" {
			userActions.Review.IsCurrentUser = true
		}

		return nil
	})

	// Send related posts fetch in a goroutine
	g.Go(func() error {

		var (
			taskErr error
			posts   types.Posts
		)

		// Don't cache the related posts only for the admin.
		if data.CurrentUser.IsAdmin(s.config) {
			posts, taskErr = s.postsRepo.GetRelatedPosts(r.Context(), post.GetTitle())
		} else {
			posts, taskErr = rdb.GetCachedData(
				r.Context(),
				s.rdb,
				fmt.Sprintf(relatedPostsCacheKey, videoID),
				s.config.CacheTimeout,
				func() (types.Posts, error) {
					return s.postsRepo.GetRelatedPosts(r.Context(), post.GetTitle())
				},
			)
		}

		// If error just log it, no related posts will be shown.
		if taskErr != nil {
			slog.ErrorContext(
				r.Context(), "failed to get related posts",
				"error", taskErr,
			)
		}

		relatedPosts = posts.Items
		return nil
	})

	// Wait for the goroutines to finish
	if err := g.Wait(); err != nil {

		slog.ErrorContext(
			r.Context(), "failed to get a single post",
			"error", err,
		)

		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Attach the results to the post object
	post.Reviews = &postReviews
	post.UserActions = userActions
	post.RelatedPosts = relatedPosts

	// Assign the post to data
	data.CurrentPost = &post
	s.ui.RenderHTML(w, r, "post.html", data)
}
