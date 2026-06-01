package worker

import (
	"fmt"
	"strings"
)

type WorkerStats struct {
	FetchedDbSources  int
	FetchedYtSources  int
	FetchedYtChannels int
	UpdatedDbSources  int
	FetchedDbVideos   int
	FetchedYtVideos   int
	AdoptedDbVideos   int
	DeletedDbVideos   []string
	InsertedDbVideos  int
	UpdatedDbVideos   int
}

func (ws WorkerStats) String() string {

	var sb strings.Builder

	fmt.Fprintf(&sb, "Fetched playlists from DB: %d\n", ws.FetchedDbSources)
	fmt.Fprintf(&sb, "Fetched playlists from YouTube: %d\n", ws.FetchedYtSources)
	fmt.Fprintf(&sb, "Fetched channels from YouTube: %d\n", ws.FetchedYtChannels)
	fmt.Fprintf(&sb, "Updated playlists in DB: %d\n", ws.UpdatedDbSources)
	fmt.Fprintf(&sb, "Fetched videos from DB: %d\n", ws.FetchedDbVideos)
	fmt.Fprintf(&sb, "Fetched valid videos from YouTube: %d\n", ws.FetchedYtVideos)
	fmt.Fprintf(&sb, "Adopted videos in DB: %d\n", ws.AdoptedDbVideos)
	fmt.Fprintf(&sb, "Deleted videos in DB: %d. %v\n", len(ws.DeletedDbVideos), ws.DeletedDbVideos)

	if len(ws.DeletedDbVideos) > 0 && len(ws.DeletedDbVideos) >= deleteLimit {
		sb.WriteString(
			"WARNING: HIT MAX DELETION LIMIT." +
				"If this persists investigate for bugs.\n",
		)
	}

	fmt.Fprintf(&sb, "Added videos in DB: %d\n", ws.InsertedDbVideos)
	fmt.Fprintf(&sb, "Updated videos in DB: %d", ws.UpdatedDbVideos)

	return sb.String()
}
