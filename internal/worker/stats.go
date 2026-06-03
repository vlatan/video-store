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

func (ws WorkerStats) Log() {

	stats := []stat{
		{"Fetched playlists from DB", ws.FetchedDbSources},
		{"Fetched playlists from YT", ws.FetchedYtSources},
		{"Fetched channels from YT", ws.FetchedYtChannels},
	}

	if ws.UpdatedDbSources > 0 {
		stats = append(stats, stat{"Updated playlists in DB", ws.UpdatedDbSources})
	}

	stats = append(stats, stat{"Fetched videos from DB", ws.FetchedDbVideos})
	stats = append(stats, stat{"Fetched videos from YT", ws.FetchedYtVideos})

	if ws.AdoptedDbVideos > 0 {
		stats = append(stats, stat{"Adopted videos in DB", ws.AdoptedDbVideos})
	}

	if len(ws.DeletedDbVideos) > 0 {
		stats = append(stats, stat{"Deleted videos in DB", len(ws.DeletedDbVideos)})
		stats = append(stats, stat{"Deleted videos ids", ws.DeletedDbVideos})
	}

	if ws.InsertedDbVideos > 0 {
		stats = append(stats, stat{"Added videos in DB", ws.InsertedDbVideos})
	}

	if ws.UpdatedDbVideos > 0 {
		stats = append(stats, stat{"Updated videos in DB", ws.UpdatedDbVideos})
	}

	logStats(stats)
}

func logStats(stats []stat) {
	maxLabel := 0
	for _, s := range stats {
		maxLabel = max(maxLabel, len(s.label))
	}

	for _, s := range stats {
		log.Printf("%-*s %v", maxLabel+1, s.label+":", s.value)
	}
}
