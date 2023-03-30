package streamingService

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

// TODO: Rename this to streaming service
// MusicPlayer is an interface that defines the methods that a music player should have.
type StreamingService interface {
  // Play starts playback of the current song.
  Play() error
  // Pause pauses playback of the current song.
  Pause() error
  // Next skips to the next song in the queue.
  Next() error
  QueueSong(song string) error
  // CurrentSong returns the current song that is playing.
  CurrentSong() (*model.CurrentlyPlayingSong, bool, error)
  TimeRemaining() (int, error)
  GetPlaylists() ([]*model.Playlist, error)
  GetSongsInPlaylist(string) ([]*model.SimpleSong, error)
  Search(string) ([]*model.SimpleSong, error)
}
