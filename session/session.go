package session

import (
	"fmt"
	"sync"
	"time"


	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
)

type Session struct {
	SessionInfo 		*model.SessionInfo
	Channels 			[]chan *model.SessionInfo
	Voters 				map[string] *voter.Voter
	MusicPlayer			musicplayer.MusicPlayer
	ChannelMutex 		*sync.Mutex
	QueueMutex   		*sync.Mutex
	VotersMutex 		*sync.Mutex
}

func NewSession() Session {
	session := Session{
		SessionInfo: 		nil,
		Channels: 			nil,
		Voters: 			make(map[string]*voter.Voter),	
		MusicPlayer: 		nil,
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
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := s.MusicPlayer.CurrentSong()
		if err != nil {
			utils.LogErrorObject("Error getting music player state", err)
			continue
		}

		if spotifyCurrentlyPlaying == true {
			if s.SessionInfo.CurrentlyPlaying.ID != spotifyCurrentlyPlayingSong.ID {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				s.SessionInfo.CurrentlyPlaying = spotifyCurrentlyPlayingSong
				sendUpdateFlag = true
				addNextSongFlag = true
			} else if s.SessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				// If same song is paused and then played, set the new state
				s.SessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft, err := s.MusicPlayer.TimeRemaining()
			if err != nil {
				utils.LogErrorObject("Error getting song time remaining", err)
				continue
			}

			if timeLeft < 5000 && addNextSongFlag {
				s.AdvanceQueue(false)

				sendUpdateFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if s.SessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				s.SessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}
		}

		if sendUpdateFlag {
			s.SendUpdate()
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (s *Session) AdvanceQueue(force bool) error {
	var song *model.Song

	s.QueueMutex.Lock()
	if len(s.SessionInfo.Queue) != 0 {
		song, s.SessionInfo.Queue = s.SessionInfo.Queue[0], s.SessionInfo.Queue[1:]
		s.QueueMutex.Unlock()

		s.MusicPlayer.QueueSong(song.ID)

	} else {
		// This else block is so we can unlock right after we update the queue in the true condition
		s.QueueMutex.Unlock()
	}

	if force {
		s.MusicPlayer.Next()
	}

	return nil
}

func (s *Session) SendUpdate() {
	go func() {
		var activeChannels []chan *model.SessionInfo

		s.ChannelMutex.Lock()
		channels := s.Channels

		for _, ch := range channels {
			select {
				case ch <- s.SessionInfo: // This is the actual send.
					// Our message went through, do nothing
					activeChannels = append(activeChannels, ch)
				default: // This is run when our send does not work.
					fmt.Println("Channel closed in update.")
					// You can handle any deregistration of the channel here.
			}
		}

		s.Channels = activeChannels
		s.ChannelMutex.Unlock()
	}()
}

// TODO: Write function that watches voters and removes any inactive ones