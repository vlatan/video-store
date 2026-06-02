package worker

import (
	"log"
)

type WorkerStats struct {
	FetchedDbSources  int
	FetchedYtSources  int
	FetchedYtChannels int
	UpdatedDbSources  int64
	FetchedDbVideos   int
	FetchedYtVideos   int
	AdoptedDbVideos   int
	DeletedDbVideos   []string
	InsertedDbVideos  int
	UpdatedDbVideos   int
}

func (ws WorkerStats) Log() {

	log.Printf("Fetched playlists from DB: %d", ws.FetchedDbSources)
	log.Printf("Fetched playlists from YouTube: %d", ws.FetchedYtSources)
	log.Printf("Fetched channels from YouTube: %d", ws.FetchedYtChannels)
	log.Printf("Updated playlists in DB: %d", ws.UpdatedDbSources)
	log.Printf("Fetched videos from DB: %d", ws.FetchedDbVideos)
	log.Printf("Fetched valid videos from YouTube: %d", ws.FetchedYtVideos)
	log.Printf("Adopted videos in DB: %d", ws.AdoptedDbVideos)
	log.Printf("Deleted videos in DB: %d; %v", len(ws.DeletedDbVideos), ws.DeletedDbVideos)

	if len(ws.DeletedDbVideos) > 0 && len(ws.DeletedDbVideos) >= deleteLimit {
		log.Println(
			"WARNING: HIT MAX DELETION LIMIT.",
			"If this persists investigate for bugs.",
		)
	}

	log.Printf("Added videos in DB: %d\n", ws.InsertedDbVideos)
	log.Printf("Updated videos in DB: %d", ws.UpdatedDbVideos)

}
