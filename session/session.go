package session

import (
	"os"
	"fmt"
	"sync"
	"time"
	"context"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"golang.org/x/exp/slog"

	"github.com/jackc/pgx/v4/pgxpool"

)

type Session struct {
	SessionInfo 		*model.SessionInfo
	Channels 			[]chan *model.SessionInfo
	Voters 				map[string] *voter.Voter
	MusicPlayer			musicplayer.MusicPlayer
	ExpiresAt			time.Time
	BonusVotes 			map[string]map[string]int
	ChannelMutex 		*sync.Mutex
	QueueMutex   		*sync.Mutex
	VotersMutex 		*sync.Mutex
	ExpiryMutex 		*sync.Mutex
	BonusVoteMutex 		*sync.Mutex
}

// Session gets removed after being inactive for this long in minutes
const sessionTimeout time.Duration = 30
// Spotify gets watched by default at this frequency in milliseconds
const spotifyWatchFrequency time.Duration = 250
// Voters get watched at this frequency in seconds
const voterWatchFrequency time.Duration = 1

func NewSession() Session {
	session := Session{
		SessionInfo: 		nil,
		Channels: 			nil,
		Voters: 			make(map[string]*voter.Voter),	
		MusicPlayer: 		nil,
		ExpiresAt: 			time.Now().Add(sessionTimeout * time.Minute),
		BonusVotes: 		make(map[string]map[string]int),
		ChannelMutex: 		&sync.Mutex{},
		QueueMutex: 		&sync.Mutex{},		
		VotersMutex: 		&sync.Mutex{},		
		ExpiryMutex: 		&sync.Mutex{},
		BonusVoteMutex: 	&sync.Mutex{},
	}

	return session
}

func (s *Session) WatchSpotifyCurrentlyPlaying() {
	s.SessionInfo.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	sendUpdateFlag := false
	addNextSongFlag := false

	for {
		if s.ExpiresAt.Before(time.Now()) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", s.SessionInfo.ID)
			return
		}

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

		// TODO: Maybe make this refresh value dynamic to adjust refresh frequency at the end of a song
		time.Sleep(spotifyWatchFrequency * time.Millisecond)
	}
}

func (s *Session) AdvanceQueue(force bool) error {
	var song *model.Song

	s.QueueMutex.Lock()
	if len(s.SessionInfo.Queue) == 0 {
		s.QueueMutex.Unlock()
		return nil
	}

	song, s.SessionInfo.Queue = s.SessionInfo.Queue[0], s.SessionInfo.Queue[1:]
	s.QueueMutex.Unlock()

	err := s.MusicPlayer.QueueSong(song.ID)
	if err != nil {
		return err
	}

	err = s.processBonusVotes(song.ID)
	if err != nil {
		return err
	}

	if !force {
		return nil
	}

	err = s.MusicPlayer.Next()
	if err != nil {
		return err
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

		s.ExpiryMutex.Lock()
		s.ExpiresAt = time.Now().Add(sessionTimeout * time.Minute)
		s.ExpiryMutex.Unlock()
	}()
}

// TODO: Write function that watches voters and removes any inactive ones
func (s *Session) WatchVoters() {

	for {
		if s.ExpiresAt.Before(time.Now()) {
			slog.Info("Session has expired, ending session voter watcher", "session_id", s.SessionInfo.ID)
			return
		}

		s.VotersMutex.Lock()
		for _, voter := range(s.Voters){
			if voter.VoterType == constants.AdminVoterType {
				continue
			}

			if voter.Expires.Before(time.Now()){
				slog.Info("Voter exipred! Removing", "voter", voter.Id)
				delete(s.Voters, voter.Id)
			}

		}
		s.VotersMutex.Unlock()

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

// TODO: This code hasn't been tested
func (s *Session) processBonusVotes(songID string) error {
	s.BonusVoteMutex.Lock()
	bonusVotes, exists := s.BonusVotes[songID]
	s.BonusVoteMutex.Unlock()
	if !exists {
		return nil
	}

	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}

	pg := database.PostgresWrapper{PostgresClient: dbPool}

	for userID, votes := range(bonusVotes){
		err = pg.SubtractBonusVotes(userID, votes)
		if err != nil {
			slog.Warn("Error updating user's bonus votes", "user", userID)
		}
	}

	pg.CloseConnection()
	return nil
}