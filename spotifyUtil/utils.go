package spotifyUtil

import (
	"io"
	"fmt"
	"strings"
	"net/http"
	"net/url"
	"encoding/base64"
	"encoding/json"
)

type Request struct {
	AccessToken string `json:"access_token"`
}

func RefreshToken(UserID int, refreshToken string) (string, error) {
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