package urls

const (
	_APIBase string = "https://api.spotify.com/v1"

	_AccountsBase string = "https://accounts.spotify.com"

	// NewToken is the  base URL of the spotify accounts API
	NewToken string = _AccountsBase + "/api/token"

	Code string = _AccountsBase + "/authorize"

	// CurrentlyPlaying is the currently-playing endpoint.
	CurrentlyPlaying string = _APIBase + "/me/player/currently-playing"

	// TrackAnalysis is the endpoint for getting the audio analysis of a track.
	TrackAnalysis string = _APIBase + "/audio-analysis"
)
