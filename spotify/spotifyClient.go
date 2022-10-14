package spotify

import "fmt"

type spotifyClient struct {
	accessToken string
}

func New (accessToken string) spotifyClient {
	s := spotifyClient{
		accessToken: accessToken,
	}
	return s
}

func (s spotifyClient) Play() {
	fmt.Printf("Playing with token: %v\n", s.accessToken)
}

func (s spotifyClient) Pause() {
	fmt.Printf("Pausing with token: %v\n", s.accessToken)
}

func (s spotifyClient) Advance(song string) {
	fmt.Printf("Advancing to %v with token: %v\n", song, s.accessToken)
}