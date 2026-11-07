package worker

import (
	"log"
)

type stat struct {
	label string
	value any
}

type WorkerStats struct {
	FetchedDbSources  int
	FetchedYtSources  int
	FetchedYtChannels int
	UpdatedDbSources  int64
	FetchedDbVideos   int
	FetchedYtVideos   int
	AdoptedDbVideos   int64
	DeletedDbVideos   []string
	InsertedDbVideos  int64
	UpdatedDbVideos   int64
}

// Log logs the worker stats
func (w *Worker) Log() {

	stats := []stat{
		{"Fetched playlists from DB", w.stats.FetchedDbSources},
		{"Fetched playlists from YT", w.stats.FetchedYtSources},
		{"Fetched channels from YT", w.stats.FetchedYtChannels},
	}

	if w.stats.UpdatedDbSources > 0 {
		stats = append(stats, stat{"Updated playlists in DB", w.stats.UpdatedDbSources})
	}

	stats = append(stats, stat{"Fetched videos from DB", w.stats.FetchedDbVideos})
	stats = append(stats, stat{"Fetched videos from YT", w.stats.FetchedYtVideos})

	if w.stats.AdoptedDbVideos > 0 {
		stats = append(stats, stat{"Adopted videos in DB", w.stats.AdoptedDbVideos})
	}

	if len(w.stats.DeletedDbVideos) > 0 {
		stats = append(stats, stat{"Deleted videos in DB", len(w.stats.DeletedDbVideos)})
		stats = append(stats, stat{"Deleted videos ids", w.stats.DeletedDbVideos})

		if len(w.stats.DeletedDbVideos) >= deleteLimit {
			stats = append(stats, stat{"WARNING: MAX DELETE LIMIT REACHED", ""})
		}
	}

	if w.stats.InsertedDbVideos > 0 {
		stats = append(stats, stat{"Added videos in DB", w.stats.InsertedDbVideos})
	}

	if w.stats.UpdatedDbVideos > 0 {
		stats = append(stats, stat{"Updated videos in DB", w.stats.UpdatedDbVideos})
	}

	logStats(stats)
}

// logStats logs the labels and values in stats
func logStats(stats []stat) {
	maxLabel := 0
	for _, s := range stats {
		maxLabel = max(maxLabel, len(s.label))
	}

	// Pad the left strings to the length of the longest string
	for _, s := range stats {
		log.Printf("%-*s %v", maxLabel+1, s.label+":", s.value)
	}
}
