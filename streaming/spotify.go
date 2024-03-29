package streaming

import (
	"context"
	"errors"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type spotifyWrapper struct {
	client *spotify.Client
}

func NewSpotifyClient(refreshToken string) *spotifyWrapper {	
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// The function spotifyauth.New() gets spotify client ID and secret from env variables, these
	// can't be read and passed manually so the names must not change
	httpClient := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(httpClient)
	return &spotifyWrapper{client: client}
}

func (s *spotifyWrapper) Play() error {
  	return s.client.Play(context.Background())
}

func (s *spotifyWrapper) Pause() error {
  	return s.client.Pause(context.Background())
}

func (s *spotifyWrapper) Next() error {
  	return s.client.Next(context.Background())
}

func (s *spotifyWrapper) QueueSong(song string) error {
	return s.client.QueueSong(context.Background(), spotify.ID(song))
}

func (s *spotifyWrapper) CurrentSong() (*model.CurrentlyPlayingSong, bool, error) {
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
		IsPlaying: status.Playing,
		SongProgressSeconds: status.Progress/1000,
		SongDurationSeconds: status.Item.Duration/1000,
	}


	return song, true, nil
}

func (s *spotifyWrapper) TimeRemaining() (int, error) {
	status, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return 0, err
	}

	if status.Item == nil {
		return 0, errors.New("currently playing song returned nil")
	}
	timeLeft := status.Item.SimpleTrack.Duration - status.Progress
	return timeLeft, nil
}

func (s *spotifyWrapper) GetPlaylists() ([]*model.Playlist, error) {
	var playlists []*model.Playlist
	currentlyPlaying, err := s.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return nil, err
	}

	playerContext := currentlyPlaying.PlaybackContext
	if playerContext.Type == "playlist" {
		slog.Debug("Getting information for currently playing playlist", "playlist-URI", playerContext.URI)
		// Kind of a hack, but can't find an easy way to convert URI into spotify ID besides trimming the string
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

	// This request only gets 20 playlists
	userPlaylists, err := s.client.CurrentUsersPlaylists(context.Background())
	if err != nil {
		return nil, err
	}

	for _, playlist := range(userPlaylists.Playlists) {
		p := model.Playlist{
			ID: string(playlist.ID),
			Name: playlist.Name,
			Image: playlist.Images[0].URL,
		}
		playlists = append(playlists, &p)
	}

	return playlists, nil
}

func (s *spotifyWrapper) GetSongsInPlaylist(playlist string) ([]*model.SimpleSong, error) {
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

func (s *spotifyWrapper) Search(query string) ([]*model.SimpleSong, error){
	searchResult, err := s.client.Search(context.Background(), query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}
	
	numberSongsToReturn := 5
	if len(searchResult.Tracks.Tracks) < 5 {
		numberSongsToReturn = len(searchResult.Tracks.Tracks)
	}
	var simpleSongList []*model.SimpleSong
	for i := 0; i < numberSongsToReturn; i++ {
		track := searchResult.Tracks.Tracks[i]
		song := SpotifyFullTrackToSimpleSong(&track)
		simpleSongList = append(simpleSongList, song)
	}

	return simpleSongList, nil
}

func SpotifyFullTrackToSimpleSong(track *spotify.FullTrack) *model.SimpleSong {
	artists := track.Artists[0].Name

	if len(track.Artists) > 1 {
		for _, artist := range(track.Artists[1:]) {
			artists += ", "
			artists += artist.Name
		}
	}

	song := &model.SimpleSong{
		ID: track.ID.String(),
		Title: track.Name,
		Artist: artists,
		Image: track.Album.Images[0].URL,
	}
	return song
}