package streaming

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
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

// SpotifyClient is a struct that implements the MusicPlayer interface for Spotify.
type SpotifyWrapper struct {
	client *spotify.Client
}

func NewSpotifyClient(accessToken string) *SpotifyWrapper {	
	// TODO: Use refresh token as well? https://pkg.go.dev/golang.org/x/oauth2#Token
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	httpClient := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(httpClient)
	return &SpotifyWrapper{client: client}
}

func (s *SpotifyWrapper) Play() error {
  	return s.client.Play(context.Background())
}

func (s *SpotifyWrapper) Pause() error {
  	return s.client.Pause(context.Background())
}

func (s *SpotifyWrapper) Next() error {
  	return s.client.Next(context.Background())
}

func (s *SpotifyWrapper) QueueSong(song string) error {
	return s.client.QueueSong(context.Background(), spotify.ID(song))
}

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

	simpleSong := SpotifyFullTrackToSimpleSong(status.Item)
	song := &model.CurrentlyPlayingSong{
		SimpleSong: simpleSong,
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

func (s *SpotifyWrapper) GetPlaylists() ([]*model.Playlist, error) {
	var playlists []*model.Playlist
	currentlyPlaying, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return nil, err
	}

	playerContext := currentlyPlaying.PlaybackContext
	if playerContext.Type == "playlist" {
		slog.Info("Getting information for currently playing playlist", "playlist-URI", playerContext.URI)
		// Kind of a hack, but can't find an easy to convert URI into spotify ID besides trimming the string
		currentPlaylist, err := s.client.GetPlaylist(context.Background(), spotify.ID(playerContext.URI[17:]))
		if err != nil {
			slog.Warn("Error getting information for currently playing spotify playlist", "error", err)
		} else {
			p := model.Playlist{
				ID: string(currentPlaylist.ID),
				Name: currentPlaylist.Name,
				Image: currentPlaylist.Images[0].URL,
			}
			playlists = append(playlists, &p)
		}
	}

	userPlaylists, err := s.client.CurrentUsersPlaylists(context.Background())
	if err != nil {
		return nil, err
	}

	for _, playlist := range(userPlaylists.Playlists[:8-len(playlists)]) {
		p := model.Playlist{
			ID: string(playlist.ID),
			Name: playlist.Name,
			Image: playlist.Images[0].URL,
		}
		playlists = append(playlists, &p)
	}

	return playlists, nil
}

func (s *SpotifyWrapper) GetSongsInPlaylist(playlist string) ([]*model.SimpleSong, error) {
	playlistItems, err := s.client.GetPlaylist(context.Background(), spotify.ID(playlist))
	if err != nil {
		return nil, err
	}

	var songs []*model.SimpleSong

	for _, song := range(playlistItems.Tracks.Tracks) {
		s := SpotifyFullTrackToSimpleSong(&song.Track)
		songs = append(songs, s)
	}

	return songs, nil
}

func (s *SpotifyWrapper) Search(query string) ([]*model.SimpleSong, error){
	searchResult, err := s.client.Search(context.Background(), query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}
	
	var simpleSongList []*model.SimpleSong
	for i := 1; i <= 5; i++ {
		track := searchResult.Tracks.Tracks[i]
		song := SpotifyFullTrackToSimpleSong(&track)
		simpleSongList = append(simpleSongList, song)
	}

	return simpleSongList, nil
}

func SpotifyFullTrackToSimpleSong(track *spotify.FullTrack) *model.SimpleSong {
	song := &model.SimpleSong{
		ID: track.ID.String(),
		Title: track.Name,
		// TODO: Loop through all artists and combine
		Artist: track.Artists[0].Name,
		Image: track.Album.Images[0].URL,
	}
	return song
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