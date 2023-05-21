//go:generate mockgen -destination=../mocks/mock_streaming.go -package=mocks . StreamingService

package streaming

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type StreamingService interface {
  Play() error
  Pause() error
  Next() error
  QueueSong(song string) error
  CurrentSong() (*model.CurrentlyPlayingSong, bool, error)
  TimeRemaining() (int, error)
  GetPlaylists() ([]*model.Playlist, error)
  GetSongsInPlaylist(string) ([]*model.SimpleSong, error)
  Search(string) ([]*model.SimpleSong, error)
}
