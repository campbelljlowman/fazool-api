package musicplayer

import (
	"context"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// SpotifyClient is a struct that implements the MusicPlayer interface for Spotify.
type SpotifyWrapper struct {
	client *spotify.Client
}

// NewSpotifyClient creates a new SpotifyClient.
func NewSpotifyClient(accessToken string) *SpotifyWrapper {	
	// TODO: Use refresh token as well? https://pkg.go.dev/golang.org/x/oauth2#Token
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	httpClient := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(httpClient)
	return &SpotifyWrapper{client: client}
}

// Play starts playback of the current song.
func (s *SpotifyWrapper) Play() error {
  	return s.client.Play(context.Background())
}

// Pause pauses playback of the current song.
func (s *SpotifyWrapper) Pause() error {
  	return s.client.Pause(context.Background())
}

// Next skips to the next song in the queue.
func (s *SpotifyWrapper) Next() error {
  	return s.client.Next(context.Background())
}

func (s *SpotifyWrapper) QueueSong(song string) error {
	return s.client.QueueSong(context.Background(), spotify.ID(song))
}

// CurrentSong returns the current song that is playing.
func (s *SpotifyWrapper) CurrentSong() (*model.CurrentlyPlayingSong, error) {
	status, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return nil, err
	}

	song := &model.CurrentlyPlayingSong{
		ID: status.Item.ID.String(),
		Title: status.Item.Name,
		// TODO: Loop through all artists and combine
		Artist: status.Item.Artists[0].Name,
		Image: status.Item.Album.Images[0].URL,
		Playing: status.Playing,
	}

	return song, nil
}

func (s *SpotifyWrapper) TimeRemaining() (int, error) {
	status, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return 0, err
	}

	timeLeft := status.Item.SimpleTrack.Duration - status.Progress
	return timeLeft, nil
}
