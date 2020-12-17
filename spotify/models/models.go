package models

// CurrentlyPlaying is the model to contain the response from the spotify currently-playing endpoint.
type CurrentlyPlaying struct {
	Timestamp int `json:"timestamp"`   // The time we made the request.
	Progress  int `json:"progress_ms"` // How far through the song we are.
	Item      struct {
		Duration int    `json:"duration_ms"` // The duration of the song.
		ID       string `json:"id"`          // The Spotify ID of the song.
		Name     string `json:"name"`
	} `json:"item"`
}

// TrackAnalysis is the model to hold all the track analysis data.
type TrackAnalysis struct {
	Beats  []timeInterval `json:"beats"`  // All the beats in track.
	Bars   []timeInterval `json:"bars"`   // All the bars in the track.
	Tatums []timeInterval `json:"tatums"` //All the tatums in the track.
	Track  struct {
		Duration float64 `json:"duration"` // The duration of the track.
	} `json:"track"`
}

type SpotifyToken struct {
	Refresh string `json:"refresh_token"`
	Access  string `json:"access_token"`
}

type timeInterval struct {
	Start    float64 `json:"start"`    // The start of the interval.
	Duration float64 `json:"duration"` // The duration of the interval.
}
