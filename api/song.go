package api

// Basic info for displaying a song in UI
type Song struct {
	SongID int `json:"song_id"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Votes int `json:"votes"`
	Image string `json:"image"`
}

type SongAction struct {
	Action string `json:"action"`
}

func PopulateSongs(){
	songs[1] = Song{
		SongID: 1,
		Title: "The Jackie",
		Artist: "J Cole",
		Votes: 0,
	}
	songs[2] = Song{
		SongID: 2,
		Title: "Myron",
		Artist: "Lil Uzi Vert",
		Votes: 0,
	}
	songs[3] = Song{
		SongID: 3,
		Title: "Wagon Wheel",
		Artist: "Darius Rucker",
		Votes: 0,
	}
}