package spotify

import "fmt"

type SpotifyClient struct {
	accessToken string
}

func New (accessToken string) SpotifyClient {
	s := SpotifyClient{
		accessToken: accessToken,
	}
	return s
}

// Error for when token is expired:     "error": {
//     "status": 401,
//     "message": "The access token expired"
// }
func (s SpotifyClient) Play() {
	fmt.Printf("Playing with token: %v\n", s.accessToken)
}

func (s SpotifyClient) Pause() {
	fmt.Printf("Pausing with token: %v\n", s.accessToken)
}

func (s SpotifyClient) Advance(song string) {
	fmt.Printf("Advancing to %v with token: %v\n", song, s.accessToken)
}