package session

import (
	"fmt"
	"sync"
	"time"
	"context"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/zmb3/spotify/v2"
)

type Session struct {
	SessionInfo 	*model.SessionInfo
	Channels 		[]chan *model.SessionInfo
	// TODO: Make this map of interfaces
	// musicPlayers		map[int] *musicplayer.MusicPlayer
	SpotifyPlayer 	*spotify.Client
	ChannelMutex 	*sync.Mutex
	QueueMutex   	*sync.Mutex
}

func NewSession () Session {
	session := Session{
		SessionInfo: 		nil,
		Channels: 			nil,
		SpotifyPlayer: 	nil,
		ChannelMutex: 		&sync.Mutex{},
		QueueMutex: 		&sync.Mutex{},		
	}

	return session
}

func (*Session) WatchSpotifyCurrentlyPlaying () {
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
			if timeLeft < 5000 && addNextSong {
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

func (* Session) sendUpdate(r *mutationResolver, sessionID int) {
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