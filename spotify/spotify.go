package spotify

import (
	"encoding/json"
	"io"
	"log"
	"main/spotify/models"
	"main/spotify/urls"
	"net/http"
	"time"
)

// type spotifyToken struct {
// 	access  string
// 	refresh string
// }

const spotifyToken = "BQBvptACLb4VAwO0MX5V93A1p1k06k_yoYRuNOLt7GEvwLnqP3lsFgYg-CTsrH5VJzibfGfP_Llok3krr4xEZQPIdLCIGETWT9yfgpmrA30yR6Kf1VNHiJTZBLVzxM0mbZv5RjmJVXQ8voFo3ywu9FgQr_Z4bWlipII"

// const tokenFile = "tokens.json"

// buildRequest returns a request that has the spotify token stored in the params.
func buildRequest(method string, url string, body io.Reader) (*http.Client, *http.Request, error) {
	client := &http.Client{
		Timeout: time.Second * 10, // 10 second timeout.
	}

	req, err := http.NewRequest(method, urls.APIBase+url, body)
	if err != nil {
		return client, req, err
	}
	// Add the access token to the request.
	req.Header.Set("Authorization", "Bearer "+spotifyToken)
	// log.Println(req.Header.Values("Authorization")[0])

	// Return the request object.
	return client, req, err
}

// GetTrackAnalysis fetches the spotify audio analysis of the supplied track.
func GetTrackAnalysis(trackID string) models.TrackAnalysis {

	client, req, err := buildRequest("GET", urls.TrackAnalysis+"/"+trackID, nil)

	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)

	// Error checks
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != 200 {
		log.Fatal(res.Status)
	}

	// Decode the data.
	var trackAn models.TrackAnalysis
	if err := json.NewDecoder(res.Body).Decode(&trackAn); err != nil {
		log.Fatal(err)
	}
	return trackAn
}

// GetCurrentlyPlaying gets the currently-playing media from spotify.
func GetCurrentlyPlaying() models.CurrentlyPlaying {

	// Make the request.
	client, req, err := buildRequest("GET", urls.CurrentlyPlaying, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)

	// Error checks
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != 200 {
		log.Fatal(res.Status)
	}

	// Decode the data.
	var currPlay models.CurrentlyPlaying
	if err := json.NewDecoder(res.Body).Decode(&currPlay); err != nil {
		log.Fatal(err)
	}
	return currPlay

}

// func loadTokens() {
// 	jsonFile, err := os.Open(tokenFile)
// 	if err != nil {
// 		fmt.
// 	}
// }
