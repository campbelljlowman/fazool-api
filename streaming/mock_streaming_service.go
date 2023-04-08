package streaming

import "github.com/campbelljlowman/fazool-api/graph/model"


type MockStreamingService struct {}

func NewMockStreamingServiceClient(accessToken string) *MockStreamingService {	
	return &MockStreamingService{}
}

func (m *MockStreamingService) Play() error {
	return nil
}

func (m *MockStreamingService) Pause() error {
	return nil
}

func (m *MockStreamingService) Next() error {
	return nil
}

func (m *MockStreamingService) QueueSong(song string) error {
	return nil
}

func (m *MockStreamingService) CurrentSong() (*model.CurrentlyPlayingSong, bool, error) {
	fakeSong := &model.CurrentlyPlayingSong{
		SimpleSong: &model.SimpleSong{
			ID: "1234",
			Title: "Fake song",
			Artist: "Campbell Lowman",
			Image: "Not an image link",
		},
		Playing: false,
	}
	return fakeSong, false, nil
}

func TimeRemaining() (int, error) {
	return 1234, nil
}

func (m *MockStreamingService) GetPlaylists() ([]*model.Playlist, error) {
	var fakePlaylists []*model.Playlist
	
	fakePlaylist := &model.Playlist{
		ID: "1234",
		Name: "Fake Playlist",
		Image: "Not an image",
	}
	fakePlaylists = append(fakePlaylists, fakePlaylist)
	return fakePlaylists, nil
}

func (m *MockStreamingService) GetSongsInPlaylist(string) ([]*model.SimpleSong, error) {
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

func (m *MockStreamingService) Search(string) ([]*model.SimpleSong, error) {
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
