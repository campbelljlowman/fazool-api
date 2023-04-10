package streaming

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type StreamingService interface {
  Play() error
  Pause() error
  Next() error
  QueueSong(song string) error
  // TODO: Remove the bool from this function since it's part of the currently playing song
  CurrentSong() (*model.CurrentlyPlayingSong, bool, error)
  TimeRemaining() (int, error)
  GetPlaylists() ([]*model.Playlist, error)
  GetSongsInPlaylist(string) ([]*model.SimpleSong, error)
  Search(string) ([]*model.SimpleSong, error)
}