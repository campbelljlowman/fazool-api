package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/slog"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/voter"

	"github.com/zmb3/spotify/v2"
)

type Session struct {
	SessionInfo 	*model.SessionInfo
	Channels 		[]chan *model.SessionInfo
	Voters 			map[string] *voter.Voter
	// TODO: Make this map of interfaces
	// musicPlayers		map[int] *musicplayer.MusicPlayer
	SpotifyPlayer 	*spotify.Client
	ChannelMutex 	*sync.Mutex
	QueueMutex   	*sync.Mutex
	VotersMutex 	*sync.Mutex
}

func NewSession() Session {
	session := Session{
		SessionInfo: 		nil,
		Channels: 			nil,
		Voters: 			make(map[string]*voter.Voter),	
		SpotifyPlayer: 		nil,
		ChannelMutex: 		&sync.Mutex{},
		QueueMutex: 		&sync.Mutex{},		
		VotersMutex: 		&sync.Mutex{},		
	}

	return session
}

func (s *Session) WatchSpotifyCurrentlyPlaying() {
	// TODO: Add logic to adjust the refresh frequency based on where in the song it is
	s.SessionInfo.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	sendUpdateFlag := false
	addNextSongFlag := false

	for {
		sendUpdateFlag = false
		playerState, err := s.SpotifyPlayer.PlayerState(context.Background())
		if err != nil {
			slog.Warn("Error getting Spotify player state", "error", err)
			continue
		}

		if playerState.CurrentlyPlaying.Playing == true {
			if s.SessionInfo.CurrentlyPlaying.ID != playerState.CurrentlyPlaying.Item.ID.String() {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				s.SessionInfo.CurrentlyPlaying.ID = playerState.CurrentlyPlaying.Item.ID.String()
				s.SessionInfo.CurrentlyPlaying.Title = playerState.CurrentlyPlaying.Item.Name
				// TODO: Loop through all artists and combine
				s.SessionInfo.CurrentlyPlaying.Artist = playerState.CurrentlyPlaying.Item.Artists[0].Name
				s.SessionInfo.CurrentlyPlaying.Image = playerState.CurrentlyPlaying.Item.Album.Images[0].URL
				s.SessionInfo.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdateFlag = true
				addNextSongFlag = true
			} else if s.SessionInfo.CurrentlyPlaying.Playing != playerState.CurrentlyPlaying.Playing {
				// If same song is paused and then played, get the new state
				s.SessionInfo.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdateFlag = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft := playerState.CurrentlyPlaying.Item.SimpleTrack.Duration - playerState.CurrentlyPlaying.Progress
			if timeLeft < 5000 && addNextSongFlag {
				s.AdvanceQueue(false)

				sendUpdateFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if s.SessionInfo.CurrentlyPlaying.Playing != playerState.CurrentlyPlaying.Playing {
				s.SessionInfo.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdateFlag = true
			}
		}

		if sendUpdateFlag {
			s.SendUpdate()
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (s *Session) AdvanceQueue(force bool) {
	var song *model.Song

	s.QueueMutex.Lock()
	if len(s.SessionInfo.Queue) != 0 {
		song, s.SessionInfo.Queue = s.SessionInfo.Queue[0], s.SessionInfo.Queue[1:]
		s.QueueMutex.Unlock()

		s.SpotifyPlayer.QueueSong(context.Background(), spotify.ID(song.ID))

	} else {
		// This else block is so we can unlock right after we update the queue in the true condition
		s.QueueMutex.Unlock()
	}

	if force {
		s.SpotifyPlayer.Next(context.Background())
	}
}

func (s *Session) SendUpdate() {
	go func() {
		s.ChannelMutex.Lock()
		channels := s.Channels
		s.ChannelMutex.Unlock()
		for _, ch := range channels {
			select {
			case ch <- s.SessionInfo: // This is the actual send.
				// Our message went through, do nothing
			default: // This is run when our send does not work.
				fmt.Println("Channel closed in update.")
				// You can handle any deregistration of the channel here.
				// TODO: remove channel from channels list if send fails
			}
		}
	}()
}

// TODO: Write function that watches voters and removes any inactive ones