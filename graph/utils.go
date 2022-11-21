package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/model"
	spotify "github.com/zmb3/spotify/v2"
)

func watchSpotifyCurrentlyPlaying(r *mutationResolver, sessionID int) {
	client := r.spotifyPlayers[sessionID]
	session := r.sessions[sessionID]
	session.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	sendUpdateFlag := false
	addNextSong := false

	for {
		sendUpdateFlag = false
		playerState, err := client.PlayerState(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}

		if playerState.CurrentlyPlaying.Playing == true {
			if session.CurrentlyPlaying.ID != playerState.CurrentlyPlaying.Item.ID.String() {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				session.CurrentlyPlaying.ID = playerState.CurrentlyPlaying.Item.ID.String()
				session.CurrentlyPlaying.Title = playerState.CurrentlyPlaying.Item.Name
				session.CurrentlyPlaying.Artist = playerState.CurrentlyPlaying.Item.Artists[0].Name
				session.CurrentlyPlaying.Image = playerState.CurrentlyPlaying.Item.Album.Images[0].URL
				session.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdateFlag = true
				addNextSong = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's 
			// Safe to add song
			timeLeft := playerState.CurrentlyPlaying.Item.SimpleTrack.Duration - playerState.CurrentlyPlaying.Progress
			if timeLeft < 5000 && addNextSong{
				advanceQueue(&r.queueMutex, session, client)

				sendUpdateFlag = true
				addNextSong = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if session.CurrentlyPlaying.Playing != playerState.CurrentlyPlaying.Playing {
				session.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdateFlag = true
			}
		}

		if sendUpdateFlag {
			sendUpdate(r, sessionID)
		}

		time.Sleep(time.Second)
	}
}

func advanceQueue(mutex *sync.Mutex, session *model.Session, client *spotify.Client) {
	var song *model.Song

	mutex.Lock()
	if len(session.Queue) != 0{
		song, session.Queue = session.Queue[0], session.Queue[1:]
		mutex.Unlock()

		client.QueueSong(context.Background(), spotify.ID(song.ID))


	} else {
		// This else block is so we can unlock right after we update the queue in the true condition
		mutex.Unlock()
	}
}

func sendUpdate (r *mutationResolver, sessionID int) {
	session := r.sessions[sessionID]
	go func() {
		r.channelMutex.Lock()
		channels := r.channels[sessionID]
		r.channelMutex.Unlock()
		for _, ch := range channels {
			select {
			case ch <- session: // This is the actual send.
				// Our message went through, do nothing
			default: // This is run when our send does not work.
				fmt.Println("Channel closed in update.")
				// You can handle any deregistration of the channel here.
			}
		}
	}()
}