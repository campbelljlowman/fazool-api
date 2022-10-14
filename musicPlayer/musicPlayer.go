package musicplayer

type MusicPlayer interface {
	play()
	pause()
	advance(song string)
}