package session

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/voter"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Session struct {
	SessionInfo    *model.SessionInfo
	channels       []chan *model.SessionInfo
	voters         map[string]*voter.Voter
	MusicPlayer    musicplayer.MusicPlayer
	expiresAt      time.Time
	// Map of [song][user][votes]
	bonusVotes     map[string]map[string]int
	queueMutex     *sync.Mutex
	channelMutex   *sync.Mutex
	votersMutex    *sync.Mutex
	expiryMutex    *sync.Mutex
	bonusVoteMutex *sync.Mutex
}

// Session gets removed after being inactive for this long in minutes
const sessionTimeout time.Duration = 30

// Spotify gets watched by default at this frequency in milliseconds
const spotifyWatchFrequency time.Duration = 250

// Voters get watched at this frequency in seconds
const voterWatchFrequency time.Duration = 1

func NewSession() Session {
	session := Session{
		SessionInfo:    nil,
		channels:       nil,
		voters:         make(map[string]*voter.Voter),
		MusicPlayer:    nil,
		expiresAt:      time.Now().Add(sessionTimeout * time.Minute),
		bonusVotes:     make(map[string]map[string]int),
		channelMutex:   &sync.Mutex{},
		queueMutex:     &sync.Mutex{},
		votersMutex:    &sync.Mutex{},
		expiryMutex:    &sync.Mutex{},
		bonusVoteMutex: &sync.Mutex{},
	}

	return session
}

func (s *Session) WatchSpotifyCurrentlyPlaying() {
	// s.SessionInfo.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	sendUpdateFlag := false
	addNextSongFlag := false

	for {
		if time.Now().After(s.expiresAt) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", s.SessionInfo.ID)
			return
		}

		sendUpdateFlag = false
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := s.MusicPlayer.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}

		if spotifyCurrentlyPlaying == true {
			if s.SessionInfo.CurrentlyPlaying.SimpleSong.ID != spotifyCurrentlyPlayingSong.SimpleSong.ID {
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
				slog.Warn("Error getting song time remaining", "error", err)
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
	var song *model.SimpleSong

	s.queueMutex.Lock()
	if len(s.SessionInfo.Queue) == 0 {
		s.queueMutex.Unlock()
		return nil
	}

	song, s.SessionInfo.Queue = s.SessionInfo.Queue[0].SimpleSong, s.SessionInfo.Queue[1:]
	s.queueMutex.Unlock()

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

		s.channelMutex.Lock()
		channels := s.channels

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

		s.channels = activeChannels
		s.channelMutex.Unlock()

		s.expiryMutex.Lock()
		s.expiresAt = time.Now().Add(sessionTimeout * time.Minute)
		s.expiryMutex.Unlock()
	}()
}

// TODO: Write function that watches voters and removes any inactive ones
func (s *Session) WatchVoters() {

	for {
		if time.Now().After(s.expiresAt) {
			slog.Info("Session has expired, ending session voter watcher", "session_id", s.SessionInfo.ID)
			return
		}

		s.votersMutex.Lock()
		for _, voter := range s.voters {
			if voter.VoterType == constants.AdminVoterType {
				continue
			}

			if time.Now().After(voter.ExpiresAt) {
				slog.Info("Voter exipred! Removing", "voter", voter.ID)
				delete(s.voters, voter.ID)
			}

		}
		s.votersMutex.Unlock()

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

// TODO: This code hasn't been tested
func (s *Session) processBonusVotes(songID string) error {
	s.bonusVoteMutex.Lock()
	bonusVotes, exists := s.bonusVotes[songID]
	delete(s.bonusVotes, songID)
	s.bonusVoteMutex.Unlock()
	if !exists {
		return nil
	}

	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}

	pg := database.PostgresWrapper{PostgresClient: dbPool}

	for userID, votes := range bonusVotes {
		err = pg.SubtractBonusVotes(userID, votes)
		if err != nil {
			slog.Warn("Error updating user's bonus votes", "user", userID)
		}
	}

	pg.CloseConnection()
	return nil
}

func (s *Session) AddChannel(newChannel chan *model.SessionInfo) {
	s.channelMutex.Lock()
	s.channels = append(s.channels, newChannel)
	s.channelMutex.Unlock()
}

func (s *Session) CloseChannels() {
	s.channelMutex.Lock()
	for _, ch := range s.channels {
		close(ch)
	}
	s.channelMutex.Unlock()
}

func (s *Session) SetQueue(newQueue [] *model.QueuedSong) {
	s.queueMutex.Lock()
	s.SessionInfo.Queue = newQueue
	s.queueMutex.Unlock()
}

func (s *Session) UpsertQueue(song model.SongUpdate, vote int) {
	s.queueMutex.Lock()
	idx := slices.IndexFunc(s.SessionInfo.Queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == song.ID })
	if idx == -1 {
		// add new song to queue
		newSong := &model.QueuedSong{
			SimpleSong: &model.SimpleSong{
				ID:     song.ID,
				Title:  *song.Title,
				Artist: *song.Artist,
				Image:  *song.Image,
			},
			Votes: vote,
		}
		s.SessionInfo.Queue = append(s.SessionInfo.Queue, newSong)
	} else {
		queuedSong := s.SessionInfo.Queue[idx]
		queuedSong.Votes += vote
	}

	// Sort queue
	sort.Slice(s.SessionInfo.Queue, func(i, j int) bool { return s.SessionInfo.Queue[i].Votes > s.SessionInfo.Queue[j].Votes })
	s.queueMutex.Unlock()
}

func (s *Session) AddVoter(voterID string, newVoter *voter.Voter){
	s.votersMutex.Lock()
	s.voters[voterID] = newVoter
	s.votersMutex.Unlock()
}

func (s *Session) GetVoter(voterID string) (*voter.Voter, bool){
	s.votersMutex.Lock()
	voter, exists := s.voters[voterID]
	s.votersMutex.Unlock()
	return voter, exists
} 

func (s *Session) IsFull() bool {
	isFull := false
	s.votersMutex.Lock()
	if len(s.voters) >= s.SessionInfo.Size {
		isFull = true
	}
	s.votersMutex.Unlock()
	return isFull
}

func (s *Session) ExpireSession() {
	s.expiryMutex.Lock()
	s.expiresAt = time.Now()
	s.expiryMutex.Unlock()
}

func (s *Session) IsExpired() bool {
	s.expiryMutex.Lock()
	isExpired := s.expiresAt.Before(time.Now())
	s.expiryMutex.Unlock()
	return isExpired
}

func (s *Session) AddBonusVote(songID, voterID string, vote int) {
	s.bonusVoteMutex.Lock()
	// TODO: Make bonus votes a map of voterID and amount
	if _, exists := s.bonusVotes[songID][voterID]; !exists {
		s.bonusVotes[songID] = make(map[string]int)
	}
	s.bonusVotes[songID][voterID] += vote
	s.bonusVoteMutex.Unlock()
}