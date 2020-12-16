package urls

const (
	// APIBase is the base URL of the spotify API.
	APIBase string = "https://api.spotify.com/v1"

	// CurrentlyPlaying is the currently-playing endpoint.
	CurrentlyPlaying string = "/me/player/currently-playing"

	// TrackAnalysis is the endpoint for getting the audio analysis of a track.
	TrackAnalysis string = "/audio-analysis"
)
