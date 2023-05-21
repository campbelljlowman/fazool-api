package streaming

import "github.com/campbelljlowman/fazool-api/graph/model"


type mockStreamingService struct {}

func NewMockStreamingServiceClient(accessToken string) *mockStreamingService {	
	return &mockStreamingService{}
}

func (m *mockStreamingService) Play() error {
	return nil
}

func (m *mockStreamingService) Pause() error {
	return nil
}

func (m *mockStreamingService) Next() error {
	return nil
}

func (m *mockStreamingService) QueueSong(song string) error {
	return nil
}

func (m *mockStreamingService) CurrentSong() (*model.CurrentlyPlayingSong, bool, error) {
	fakeSong := &model.CurrentlyPlayingSong{
		SimpleSong: &model.SimpleSong{
			ID: "1234",
			Title: "Fake song",
			Artist: "Campbell Lowman",
			Image: "Not an image link",
		},
		IsPlaying: false,
		SongProgressSeconds: 69,
		SongDurationSeconds: 420,
	}
	return fakeSong, false, nil
}

func (m *mockStreamingService) TimeRemaining() (int, error) {
	return 1234, nil
}

func (m *mockStreamingService) GetPlaylists() ([]*model.Playlist, error) {
	var fakePlaylists []*model.Playlist
	
	fakePlaylist := &model.Playlist{
		ID: "1234",
		Name: "Fake Playlist",
		Image: "Not an image",
	}
	fakePlaylists = append(fakePlaylists, fakePlaylist)
	return fakePlaylists, nil
}

func (m *mockStreamingService) GetSongsInPlaylist(string) ([]*model.SimpleSong, error) {
	var fakeSongs []*model.SimpleSong
	fakeSong := &model.SimpleSong{
		ID: "1234",
		Title: "Fake song",
		Artist: "Campbell Lowman",
		Image: "Not an image link",
	}
	fakeSongs = append(fakeSongs, fakeSong)
	return fakeSongs, nil
}

func (m *mockStreamingService) Search(string) ([]*model.SimpleSong, error) {
	var fakeSongs []*model.SimpleSong
	fakeSong := &model.SimpleSong{
		ID: "1234",
		Title: "Fake song",
		Artist: "Campbell Lowman",
		Image: "Not an image link",
	}
	fakeSongs = append(fakeSongs, fakeSong)
	return fakeSongs, nil
}
