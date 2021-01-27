package models

// Media is the model to contain the response from the spotify currently-playing endpoint.
type Media struct {
	Timestamp int  `json:"timestamp"`   // The time we made the request.
	Progress  int  `json:"progress_ms"` // How far through the song we are.
	IsPlaying bool `json:"is_playing"`  // Whether the song is currently playing or not.
	Item      struct {
		Duration int    `json:"duration_ms"` // The duration of the song.
		ID       string `json:"id"`          // The Spotify ID of the song.
		Name     string `json:"name"`
	} `json:"item"`
}

// MediaAudioFeatures is the model to hold all the track analysis data.
type MediaAudioFeatures struct {
	Acousticness     float64 `json:"acousticness"`
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Loudness         float64 `json:"loudness"`
	Speechiness      float64 `json:"speechiness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:tempo`
}

// MediaAudioAnalysis is the model to hold all the track analysis data.
type MediaAudioAnalysis struct {
	Beats  []TimeInterval `json:"beats"`  // All the beats in track.
	Bars   []TimeInterval `json:"bars"`   // All the bars in the track.
	Tatums []TimeInterval `json:"tatums"` //All the tatums in the track.
	Track  struct {
		Duration float64 `json:"duration"` // The duration of the track.
	} `json:"track"`
}

// SpotifyToken contains the spotify access and refresh tokens.
type SpotifyToken struct {
	Refresh string `json:"refresh_token"`
	Access  string `json:"access_token"`
}

type TimeInterval struct {
	Start    float64 `json:"start"`    // The start of the interval.
	Duration float64 `json:"duration"` // The duration of the interval.
}

type Trigger struct {
	Number   int `json:"number"`
	Duration int `json:"duration"`
}
