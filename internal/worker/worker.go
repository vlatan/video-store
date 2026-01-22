package worker

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/integrations/gemini"
	"github.com/vlatan/video-store/internal/integrations/yt"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/repositories/categories"
	"github.com/vlatan/video-store/internal/repositories/posts"
	"github.com/vlatan/video-store/internal/repositories/sources"
	"github.com/vlatan/video-store/internal/utils"

	"google.golang.org/api/youtube/v3"
)

type Worker struct {
	id          string
	rdb         *rdb.Service
	postsRepo   *posts.Repository
	sourcesRepo *sources.Repository
	catsRepo    *categories.Repository
	config      *config.Config
	youtube     *yt.Service
	gemini      *gemini.Service
}

// Maximum videos to delete per run
const deleteLimit = 5

// Redis key to lock the worker
const workerLockKey = "worker:lock"

func New() *Worker {

	// Create essential services
	cfg := config.New()

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("couldn't create DB service; %v", err)
	}

	rdb, err := rdb.New(cfg)
	if err != nil {
		log.Fatalf("couldn't create Redis service; %v", err)
	}

	// Create DB repositories
	postsRepo := posts.New(db, cfg)
	sourcesRepo := sources.New(db)
	catsRepo := categories.New(db)

	// Create YouTube service
	ctx := context.Background()
	yt, err := yt.New(ctx, cfg)
	if err != nil {
		log.Fatalf("couldn't create YouTube service: %v", err)
	}

	// Create Gemini client
	gemini, err := gemini.New(ctx, cfg, rdb)
	if err != nil {
		log.Fatalf("couldn't create Gemini service: %v", err)
	}

	return &Worker{
		id:          uuid.New().String(),
		rdb:         rdb,
		postsRepo:   postsRepo,
		sourcesRepo: sourcesRepo,
		catsRepo:    catsRepo,
		config:      cfg,
		youtube:     yt,
		gemini:      gemini,
	}
}

// Run the worker
func (w *Worker) Run(ctx context.Context) error {

	// Print separator at the end
	defer fmt.Println(strings.Repeat("-", 70))

	// Measure execution time
	start := time.Now()
	defer func() {
		elapsed := time.Since(start).Round(time.Second)
		log.Printf("Time took: %s", elapsed)
	}()

	// Create a new context tailored to this worker expected runtime
	ctx, cancel := context.WithTimeout(ctx, w.config.WorkerExpectedRuntime)
	defer cancel()

	// Create and acquire a Redis lock with slightly bigger TTL than the context
	redisLockTTL := time.Duration(float64(w.config.WorkerExpectedRuntime) * 1.25)
	lock := w.rdb.NewLock(workerLockKey, w.id, redisLockTTL)

	// This is a blocking call until the lock is acquired
	if err := lock.Lock(ctx); err != nil {
		return fmt.Errorf("failed to acquire Redis lock; %w", err)
	}

	// Make sure to unlock the worker when done.
	// Use ctx without cancel so Unlock isn't killed by the expired ctx.
	defer lock.Unlock(context.WithoutCancel(ctx))

	// Define retry configs for the external APIs
	ytRetryConfig := &utils.RetryConfig{
		MaxRetries: 3,
		MaxJitter:  time.Second,
		Delay:      time.Second,
	}

	geminiRetryConfig := &utils.RetryConfig{
		MaxRetries: 3,
		MaxJitter:  2 * time.Second,
		Delay:      65 * time.Second,
	}

	log.Println("Lock acquired!")
	log.Println("Worker running...")

	// GET ALL THE PLAYLISTS FROM DATABASE
	// ###################################################################

	// Fetch all the playlists from DB
	dbSources, err := w.sourcesRepo.GetSources(ctx)

	if err != nil || len(dbSources) == 0 {
		return fmt.Errorf(
			"could not fetch the sources from DB; rows: %v; %w",
			len(dbSources), err,
		)
	}

	items := utils.Plural(len(dbSources), "playlist")
	log.Printf("Fetched %d %s from DB", len(dbSources), items)

	// GET GIVEN PLAYLISTS FROM YOUTUBE
	// ###################################################################

	// Extract playlist IDs and create DB sources map
	dbSourcesMap := make(map[string]*models.Source, len(dbSources))
	playlistIDs := make([]string, len(dbSources))
	for i, source := range dbSources {
		dbSourcesMap[source.PlaylistID] = &source
		playlistIDs[i] = source.PlaylistID
	}

	// Fetch playlists from YouTube
	ytSources, err := w.youtube.GetSources(ctx, ytRetryConfig, playlistIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the playlists from YouTube; %w",
			err,
		)
	}

	items = utils.Plural(len(ytSources), "playlist")
	log.Printf("Fetched %d %s from YouTube", len(ytSources), items)

	// GET GIVEN CHANNELS FROM YOUTUBE
	// ###################################################################

	// Extract channel IDs and create YT sources map
	ytSourcesMap := make(map[string]*youtube.Playlist, len(ytSources))
	channelIDs := make([]string, len(ytSources))
	for i, source := range ytSources {
		ytSourcesMap[source.Id] = source
		channelIDs[i] = source.Snippet.ChannelId
	}

	// Fetch corresponding channels
	channels, err := w.youtube.GetChannels(ctx, ytRetryConfig, channelIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not fetch the channels from YouTube; %w",
			err,
		)
	}

	// Create channels map
	channelsMap := make(map[string]*youtube.Channel, len(channels))
	for _, channel := range channels {
		channelsMap[channel.Id] = channel
	}

	// UPDATE THE PLAYLISTS IN DATABASE
	// ###################################################################

	// Update each playlist in DB if change in thumbnails
	var updatedPlaylists int
	for playlistID, ytSource := range ytSourcesMap {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		newSource := w.youtube.NewYouTubeSource(
			ytSource, channelsMap[ytSource.Snippet.ChannelId],
		)

		// Check if channel thumbnails have changed
		if models.ThumbnailsEqual(
			dbSourcesMap[playlistID].ChannelThumbnails,
			newSource.ChannelThumbnails,
		) {
			continue
		}

		_, err = w.sourcesRepo.UpdateSource(ctx, newSource)
		if err == nil {
			updatedPlaylists++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"could not update source '%s' in DB; %v",
			newSource.PlaylistID, err,
		)
	}

	if updatedPlaylists > 0 {
		items = utils.Plural(updatedPlaylists, "playlist")
		log.Printf("Updated %d %s", updatedPlaylists, items)
	}

	// GET ALL THE VIDEOS FROM DATABASE
	// ###################################################################

	// Get ALL videos from DB, should be ordered by upload date
	dbVideos, err := w.postsRepo.GetAllPosts(ctx)
	if err != nil || len(dbVideos) == 0 {
		return fmt.Errorf(
			"could not fetch the videos from DB; rows: %v; %w",
			len(dbVideos), err,
		)
	}

	items = utils.Plural(len(dbVideos), "video")
	log.Printf("Fetched %d %s from DB", len(dbVideos), items)

	// GET ALL THE ORPHAN VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Collect the orphans video IDs
	var orphanVideoIDs []string
	for _, video := range dbVideos {
		if video.PlaylistID == "" {
			orphanVideoIDs = append(orphanVideoIDs, video.VideoID)
		}
	}

	// Get orphans metadata from YT, start forming valid YT videos map
	ytVideosMap := make(map[string]*models.Post)
	ytOrphanVideos, err := w.youtube.GetVideos(ctx, ytRetryConfig, orphanVideoIDs...)
	if err != nil {
		return fmt.Errorf(
			"could not get the orphan videos from YouTube; %w",
			err,
		)
	}

	// Add valid orphan videos to YT map
	for _, video := range ytOrphanVideos {
		if err = w.youtube.ValidateYouTubeVideo(video); err == nil {
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, "")
		}
	}

	// GET ALL THE PLAYLIST VALID VIDEOS FROM YOUTUBE
	// ###################################################################

	// Get valid videos from playlists
	for _, playlistID := range playlistIDs {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		sourceItems, err := w.youtube.GetSourceItems(ctx, ytRetryConfig, playlistID)
		if err != nil {
			return fmt.Errorf(
				"couldn't get items from YouTube for source '%s'; %w",
				playlistID, err,
			)
		}

		// Collect the video IDs for this source
		var videoIDs []string
		for _, item := range sourceItems {
			videoIDs = append(videoIDs, item.ContentDetails.VideoId)
		}

		// Get all the videos metadata for this source
		videosMetadata, err := w.youtube.GetVideos(ctx, ytRetryConfig, videoIDs...)
		if err != nil {
			return fmt.Errorf(
				"couldn't get videos from YouTube for source %s; %w",
				playlistID, err,
			)
		}

		// Keep only the valid videos
		for _, video := range videosMetadata {

			// Check the context first
			if err = ctx.Err(); err != nil {
				return err
			}

			// Skip if invalid video
			if err = w.youtube.ValidateYouTubeVideo(video); err != nil {
				continue
			}

			// Skip if the video is banned (manually deleted).
			// If error is nil the post is IN the deleted_post table.
			err = w.postsRepo.IsPostBanned(ctx, video.Id)
			if err == nil {
				continue
			}

			// Exit early if context ended
			if utils.IsContextErr(err) {
				return err
			}

			// If the video is already in ytVideosMap as an orphaned video
			// we overwrite it, associate it with a YT playlist
			ytVideosMap[video.Id] = w.youtube.NewYouTubePost(video, playlistID)
		}
	}

	items = utils.Plural(len(ytVideosMap), "video")
	log.Printf("Fetched %d valid %s from YouTube", len(ytVideosMap), items)

	// UPDATE VIDEOS' PLAYLIST IDS IN DATABASE
	// ###################################################################

	var adopted int
	for _, dbVideo := range dbVideos {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		// Check if DB video exists on YouTube
		ytVideo, exists := ytVideosMap[dbVideo.VideoID]
		if !exists {
			continue
		}

		// Check if we need to update the video playlist
		if ytVideo.PlaylistID == dbVideo.PlaylistID {
			continue
		}

		_, err = w.postsRepo.UpdatePlaylist(
			ctx, dbVideo.VideoID, ytVideo.PlaylistID,
		)

		if err == nil {
			adopted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"Failed to update the playlist on video '%s'; %v",
			dbVideo.VideoID, err,
		)
	}

	if adopted > 0 {
		items = utils.Plural(adopted, "video")
		log.Printf("Adopted %d %s", adopted, items)
	}

	// DELETE THE OBSOLETE VIDEOS FROM DATABASE
	// ###################################################################

	// Delete videos in DB
	var deleted int
	var deletedVideoIDs []string
	var validDBVideos []*models.Post
	for _, dbVideo := range dbVideos {

		// Check the context first
		if err = ctx.Err(); err != nil {
			return err
		}

		// If the DB video exists on YouTube keep it as valid
		if _, exists := ytVideosMap[dbVideo.VideoID]; exists {
			// Keep valid DB videos
			validDBVideos = append(validDBVideos, dbVideo)

			// Delete valid DB videos from the YT map.
			// In this map ONLY the ones that are not in the DB will remain.
			// Meaning the NEW videos that need to be added.
			delete(ytVideosMap, dbVideo.VideoID)

			continue
		}

		// Do not delete any more videos if max deletion was reached
		if deleted >= deleteLimit {
			break
		}

		// Delete the video
		_, err = w.postsRepo.DeletePost(ctx, dbVideo.VideoID)
		if err == nil {
			deletedVideoIDs = append(deletedVideoIDs, dbVideo.VideoID)
			deleted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"could not delete the video '%s' in DB; %v",
			dbVideo.VideoID, err,
		)
	}

	if deleted > 0 {
		items = utils.Plural(deleted, "video")
		log.Printf("Deleted %d %s", deleted, items)
		log.Printf("Deleted videos: %v", deletedVideoIDs)

		if deleted >= deleteLimit {
			msg := "WARNING: HIT MAX DELETION LIMIT. "
			msg += "If this persists investigate for bugs."
			log.Println(msg)
		}
	}

	// GET ALL THE CATEGORIES FROM DATABASE
	// ###################################################################

	// Get the categories
	categories, err := w.catsRepo.GetCategories(ctx)

	if err != nil || len(categories) == 0 {
		return fmt.Errorf(
			"could not fetch the categories from DB; rows: %v; %w",
			len(categories), err,
		)
	}

	// INSERT THE NEW VIDEOS IN DATABASE
	// ###################################################################

	// Put new YT videos in a slice.
	// ytVideosMap should now contain only new videos.
	newVideos := slices.Collect(maps.Values(ytVideosMap))

	// Summarize new videos in place
	_, err = w.summarizeVideos(
		ctx, lock, categories,
		geminiRetryConfig, newVideos,
	)

	if err != nil {
		return err
	}

	// Insert new videos in DB
	var inserted int
	for _, newVideo := range newVideos {

		_, err = w.postsRepo.InsertPost(ctx, newVideo)
		if err == nil {
			inserted++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"failed to insert video '%s' in DB; %v",
			newVideo.VideoID, err,
		)
	}

	if inserted > 0 {
		items = utils.Plural(inserted, "video")
		log.Printf("Added %d %s", inserted, items)
	}

	// UPDATE THE EXISTING VIDEOS IN DATABASE
	// ###################################################################

	// Summarize the existing videos in place
	indexes, err := w.summarizeVideos(
		ctx, lock, categories,
		geminiRetryConfig, validDBVideos,
	)

	if err != nil {
		return err
	}

	// Update the existing DB videos
	var updated int
	for _, index := range indexes {

		video := validDBVideos[index]
		_, err = w.postsRepo.UpdateGeneratedData(ctx, video)
		if err == nil {
			updated++
			continue
		}

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return err
		}

		log.Printf(
			"failed to update generated data in DB on video '%s'; %v",
			video.VideoID, err,
		)
	}

	if updated > 0 {
		items = utils.Plural(updated, "video")
		log.Printf("Updated %d %s", updated, items)
	}

	return nil
}

// summarizeVideos summarizes and categorizes videos in place,
// and returns ther indicies.
func (w *Worker) summarizeVideos(
	ctx context.Context,
	lock *rdb.RedisLock,
	categories models.Categories,
	rc *utils.RetryConfig,
	videos []*models.Post) ([]int, error) {

	var summarizedIndicies []int
	for i, video := range videos {

		// Check the context first
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Skip summarizing videos if daily quota was reached
		if w.gemini.Exhausted(ctx) {
			break
		}

		// UNCOMMENT
		// Nothing to update, summary and category are populated
		// if video.Summary != "" &&
		// 	video.Category != nil &&
		// 	video.Category.Name != "" {
		// 	continue
		// }

		// REMOVE
		// Nothing to update, summary and category are populated
		if strings.Contains(video.Summary, utils.UpdateMarker) &&
			video.Category != nil &&
			video.Category.Name != "" {
			continue
		}

		// Sleep 20s before making a genai request to avoid hitting the RPM quota
		if err := utils.SleepContext(ctx, 20*time.Second); err != nil {
			return nil, err
		}

		// Check if we still own the lock before an expensive API call
		if err := lock.CheckLock(ctx); err != nil {
			return nil, fmt.Errorf(
				"this worker %s does not own the lock anymore; %w",
				w.id, err,
			)
		}

		// Generate content using Gemini
		genaiResponse, err := w.gemini.Summarize(ctx, video, categories, rc)

		// Exit early if context ended
		if utils.IsContextErr(err) {
			return nil, err
		}

		// Skip the video update if error
		if err != nil {
			log.Printf(
				"gemini content generation on video '%s' failed; %v",
				video.VideoID, err,
			)
			continue
		}

		// Update the video in the given slice and record its index
		videos[i].Summary = genaiResponse.Summary
		videos[i].Category = &models.Category{Name: genaiResponse.Category}
		summarizedIndicies = append(summarizedIndicies, i)
	}

	return summarizedIndicies, nil
}
