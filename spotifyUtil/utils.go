package spotifyUtil

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"encoding/base64"
	"io"
	"encoding/json"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zmb3/spotify/v2"
)

type Request struct {
	AccessToken string `json:"access_token"`
}

func RefreshToken(db *pgxpool.Pool, UserID int) (string, error) {
	// Get refresh Token from DB
	refreshToken, err := database.GetSpotifyRefreshToken(db, UserID)
	if err != nil {
		println("Error getting spotify refresh token!")
		return "", fmt.Errorf("Got error %s", err.Error())
	}

	// Hit spotify endpoint to refresh token
	// TODO: Get these from env
	spotifyClientAuth := "a7666d8987c7487b8c8f345126bd1f0c:efa8b45e4d994eaebc25377afc5a9e8d"
	authString := fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(spotifyClientAuth)))
	urlPath := "https://accounts.spotify.com/api/token"
	
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	encodedData := data.Encode()
	req, err := http.NewRequest("POST", urlPath, strings.NewReader(encodedData))
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authString)
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Got error %s", err.Error())
	}
	tokenData := Request{}
	json.Unmarshal([]byte(body), &tokenData)

	// Add tokens back to database
	err = database.SetSpotifyAccessToken(db, UserID, tokenData.AccessToken)

	if err != nil {
		return "", err
	}
	
	return tokenData.AccessToken, nil
}

func WatchCurrentlyPlaying(session *model.Session, client *spotify.Client, channels []chan *model.Session) {
	println("Watching session: %v", session.ID)

	currentlyPlaying := &model.CurrentlyPlayingSong{
		ID: "55",
		Title: "Test Song",
		Artist: "Test Artist",
	}

	session.CurrentlyPlaying = currentlyPlaying

	for _, ch := range channels {
		select {
		case ch <- session: // This is the actual send.
			// Our message went through, do nothing
		default: // This is run when our send does not work.
			fmt.Println("Channel closed in update.")
			// You can handle any deregistration of the channel here.
		}
	}
}