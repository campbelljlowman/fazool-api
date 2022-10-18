package spotify

import (
	"fmt"
	"io"
	"net/http"
)

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
	client := &http.Client{}
	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/play", nil)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", s.accessToken))
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}

	if response.StatusCode != 200 {
		fmt.Printf("Request responsed with status code: %v", response.StatusCode)
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v", err)
		}
		bodyText := string(body)
		fmt.Printf("Error body: %v", bodyText)
	}
}

func (s SpotifyClient) Pause() {
	fmt.Printf("Pausing with token: %v\n", s.accessToken)
}

func (s SpotifyClient) Advance(song string) {
	fmt.Printf("Advancing to %v with token: %v\n", song, s.accessToken)
}