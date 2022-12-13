package musicplayer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

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

// CurrentSong returns the current song that is playing and a bool that indicates whether it's playing or not.
func (s *SpotifyWrapper) CurrentSong() (*model.CurrentlyPlayingSong, bool, error) {
	status, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return nil, false, err
	}

	if !status.Playing {
		return nil, false, nil
	} 

	if status.Item == nil {
		return nil, false, errors.New("Currently playing is set to true but no track is found!")
	}

	song := &model.CurrentlyPlayingSong{
		ID: status.Item.ID.String(),
		Title: status.Item.Name,
		// TODO: Loop through all artists and combine
		Artist: status.Item.Artists[0].Name,
		Image: status.Item.Album.Images[0].URL,
		Playing: status.Playing,
	}


	return song, true, nil
}

func (s *SpotifyWrapper) TimeRemaining() (int, error) {
	status, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return 0, err
	}

	timeLeft := status.Item.SimpleTrack.Duration - status.Progress
	return timeLeft, nil
}

type Request struct {
	AccessToken string `json:"access_token"`
}

func RefreshSpotifyToken(refreshToken string) (string, error) {
	// Hit spotify endpoint to refresh token
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyClientAuth := fmt.Sprintf("%v:%v", spotifyClientID, spotifyClientSecret)

	authString := fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(spotifyClientAuth)))
	urlPath := "https://accounts.spotify.com/api/token"
	
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	encodedData := data.Encode()
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(encodedData))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authString)
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	tokenData := Request{}
	json.Unmarshal([]byte(body), &tokenData)
	
	return tokenData.AccessToken, nil
}