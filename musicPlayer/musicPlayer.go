// package main

// import (
//   "context"
//   "fmt"
//   "log"

//   "github.com/zmb3/spotify"
// )

// // MusicPlayer is an interface that defines the methods that a music player should have.
// type MusicPlayer interface {
//   // Play starts playback of the current song.
//   Play() error
//   // Pause pauses playback of the current song.
//   Pause() error
//   // Next skips to the next song in the queue.
//   Next() error
//   // Previous skips to the previous song in the queue.
//   Previous() error
//   // CurrentSong returns the current song that is playing.
//   CurrentSong() (string, error)
// }

// // SpotifyClient is a struct that implements the MusicPlayer interface for Spotify.
// type SpotifyClient struct {
//   client *spotify.Client
// }

// // NewSpotifyClient creates a new SpotifyClient.
// func NewSpotifyClient(accessToken string) *SpotifyClient {
//   authenticator := spotify.NewAuthenticator("")
//   client := authenticator.NewClient(accessToken)

//   return &SpotifyClient{client: client}
// }

// // Play starts playback of the current song.
// func (s *SpotifyClient) Play() error {
//   return s.client.Play()
// }

// // Pause pauses playback of the current song.
// func (s *SpotifyClient) Pause() error {
//   return s.client.Pause()
// }

// // Next skips to the next song in the queue.
// func (s *SpotifyClient) Next() error {
//   return s.client.Next()
// }

// // Previous skips to the previous song in the queue.
// func (s *SpotifyClient) Previous() error {
//   return s.client.Previous()
// }

// // CurrentSong returns the current song that is playing.
// func (s *SpotifyClient) CurrentSong() (string, error) {
//   status, err := s.client.PlayerCurrentlyPlaying()
//   if err != nil {
//     return "", err
//   }

//   return status.Item.Name, nil
// }

// // AppleMusicClient is a struct that implements the MusicPlayer interface for Apple Music.
// type AppleMusicClient struct {
//   // TODO: Implement the AppleMusicClient
// }

// // NewAppleMusicClient creates a new AppleMusicClient.
// func NewAppleMusicClient() *AppleMusicClient {
//   // TODO: Implement the NewAppleMusicClient constructor
// }

// // Play starts playback of the current song.
// func (a *AppleMusicClient) Play() error {
//   // TODO: Implement the Play method
// }

// // Pause pauses playback of the current song.
// func (a *AppleMusicClient) Pause() error {
//   // TODO: Implement the Pause method
// }

// // Next skips to the next song in the queue.
// func (a *AppleMusicClient) Next() error {
//   // TODO: Implement the Next method
// }

// // Previous skips to the previous song in the queue.
// func (a *AppleMusicClient) Previous() error {
//   // TODO: Implement the Previous method
// }

// // CurrentSong returns the current song that is playing.
// func (a *AppleMusic
