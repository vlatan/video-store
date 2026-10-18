package posts

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/post"
	"golang.org/x/sync/errgroup"
)

// Handle a single post
func (s *Service) Single(w http.ResponseWriter, r *http.Request) {

	// Get video id from URL path
	videoID := r.PathValue("video")

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	var (
		p            post.Post
		postReviews  post.Reviews
		userActions  post.Actions
		relatedPosts []post.Post
		currentUser  = ctxd.GetUser(r.Context())
	)

	g := new(errgroup.Group)

	// Send post fetch in a goroutine
	g.Go(func() error {

		// Don't cache single post for logged in users
		var taskErr error
		if currentUser.IsAuthenticated() {
			p, taskErr = s.post.Single(r.Context(), videoID)
		} else {
			p, taskErr = s.rdb.Get(
				r.Context(),
				fmt.Sprintf(postCacheKey, videoID),
				s.cfg.CacheTimeout,
				func() (post.Post, error) {
					return s.post.Single(r.Context(), videoID)
				},
			)
		}

		return taskErr
	})

	// Get reviews from DB in a goroutine
	g.Go(func() error {

		// Get post reviews, don't cache the reviews for logged in users
		var taskErr error
		if currentUser.IsAuthenticated() {
			postReviews, taskErr = s.post.Reviews(r.Context(), videoID, "")
		} else {
			postReviews, taskErr = s.rdb.Get(
				r.Context(),
				fmt.Sprintf(postReviewsCacheKey, videoID),
				s.cfg.CacheTimeout,
				func() (post.Reviews, error) {
					return s.post.Reviews(r.Context(), videoID, "")
				},
			)
		}

		if taskErr != nil {
			return taskErr
		}

		// Get the user avatars
		for i, review := range postReviews.Items {
			localAvatarURL, taskErr := s.avatar.Get(r.Context(), review.PublicID, review.AvatarURL)
			if taskErr != nil {
				return taskErr
			}
			postReviews.Items[i].LocalAvatarURL = localAvatarURL
		}

		// Check if the current user owns a review
		for i, review := range postReviews.Items {
			if currentUser.IsAuthenticated() && currentUser.ID == review.UserID {
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
			s.tmpl.HTMLError(w, r, http.StatusNotFound)
			return
		}

		slog.ErrorContext(
			r.Context(), "failed to get a single post",
			"error", err,
		)

		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	// Get current user actions in a goroutine
	g.Go(func() error {

		// Check whether the current user liked, faved, rated and/or reviewed the post
		if currentUser.IsAuthenticated() {

			var taskErr error
			userActions, taskErr = s.post.Actions(
				r.Context(),
				currentUser.ID,
				p.ID,
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
			posts   post.Posts
		)

		// Don't cache the related posts only for the admin.
		if currentUser.IsAdmin(s.cfg) {
			posts, taskErr = s.post.Related(r.Context(), p.GetTitle())
		} else {
			posts, taskErr = s.rdb.Get(
				r.Context(),
				fmt.Sprintf(relatedPostsCacheKey, videoID),
				s.cfg.CacheTimeout,
				func() (post.Posts, error) {
					return s.post.Related(r.Context(), p.GetTitle())
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

		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	// Attach the results to the post object
	p.Reviews = &postReviews
	p.UserActions = userActions
	p.RelatedPosts = relatedPosts

	// Get template data
	data := postData{
		Base:        s.tmpl.Base(w, r),
		CurrentPost: &p,
	}

	s.tmpl.RenderHTML(w, r, "post.html", data)
}
