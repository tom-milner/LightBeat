package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/spotify/credentials"
	"main/spotify/models"
	"main/spotify/urls"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var spotifyToken models.SpotifyToken
var localIP net.IP

// buildAPIRequest returns a request that has the spotify token stored in the params.
func buildAPIRequest(method string, url string, body io.Reader) (*http.Client, *http.Request, error) {
	client := &http.Client{
		Timeout: time.Second * 10, // 10 second timeout.
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return client, req, err
	}
	// Add the access token to the request.
	req.Header.Set("Authorization", "Bearer "+spotifyToken.Access)
	// log.Println(req.Header.Values("Authorization")[0])

	// Return the request object.
	return client, req, err
}

// GetTrackAnalysis fetches the spotify audio analysis of the supplied track.
func GetTrackAnalysis(trackID string) models.TrackAnalysis {

	client, req, err := buildAPIRequest("GET", urls.TrackAnalysis+"/"+trackID, nil)

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
	if !(res.StatusCode < 300 && res.StatusCode > 100) {
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
	client, req, err := buildAPIRequest("GET", urls.CurrentlyPlaying, nil)
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
	if !(res.StatusCode < 300 && res.StatusCode > 100) {
		log.Fatal(res.Status)
	}

	var currPlay models.CurrentlyPlaying
	if res.StatusCode == 204 {
		return currPlay
	}

	// Decode the data.
	if err := json.NewDecoder(res.Body).Decode(&currPlay); err != nil {
		log.Fatal(err)
	}
	return currPlay
}

func saveRefreshToken(refreshToken string, tokenFile string) bool {

	log.Println("Saving Token")
	jsonFile, err := os.Create(tokenFile)
	if err != nil {
		log.Println(err)
		return false
	}
	defer jsonFile.Close()
	encoder := json.NewEncoder(jsonFile)
	if err = encoder.Encode(refreshToken); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func loadRefreshToken(tokenFile string) (string, error) {
	var refreshToken string
	jsonFile, err := os.Open(tokenFile)
	if err != nil {
		log.Println(err)
		return refreshToken, err
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)

	if err := decoder.Decode(&refreshToken); err != nil {
		log.Println(err)
		return refreshToken, err
	}
	return refreshToken, nil
}

// Authorize this layer with the spotify API using OAuth2
func Authorize(tokenFile string) bool {
	refreshToken, err := loadRefreshToken(tokenFile)
	if err != nil {
		log.Println(err)
		code, redirectURI := getAuthCode()
		spotifyToken = getRefreshAndAccessToken(code, redirectURI)
		saveRefreshToken(spotifyToken.Refresh, tokenFile)
	} else {
		log.Println("Refresh token found.")
		spotifyToken = getAccessToken(refreshToken)
	}
	return true
}

func getAccessToken(refreshToken string) models.SpotifyToken {
	log.Println("Fetching new access token")
	body := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	tokens := getSpotifyToken(body)
	tokens.Refresh = refreshToken
	return tokens
}

func getRefreshAndAccessToken(code string, redirectURI string) models.SpotifyToken {
	log.Println("Fetching new token pair.")
	body := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}

	tokens := getSpotifyToken(body)
	return tokens
}
func getSpotifyToken(body url.Values) models.SpotifyToken {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("POST", urls.NewToken, strings.NewReader(body.Encode()))
	req.SetBasicAuth(credentials.ClientID, credentials.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var tokenPair models.SpotifyToken
	if err := json.NewDecoder(res.Body).Decode(&tokenPair); err != nil {
		log.Fatal(err)
	}
	log.Println("Token pair fetched successfully")
	return tokenPair
}

func getAuthCode() (string, string) {

	// The use must visit this link to authenticate spotify for the first time.
	u, err := url.Parse(urls.Code)
	if err != nil {
		log.Fatal(err)
	}
	ip := getOutboundIP().String()
	serverPort := "8080"
	serverAddress := "http://" + ip + ":" + serverPort
	redirectURI := serverAddress + "/code"
	log.Println("Current IP: " + ip)
	q := u.Query()
	q.Set("client_id", credentials.ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "user-modify-playback-state,user-read-currently-playing,user-read-playback-state")
	u.RawQuery = q.Encode()

	// The user must click this link.
	log.Printf("\n\n%s\n\n", u)

	ctx, cancel := context.WithCancel(context.Background())
	mux := http.NewServeMux()

	var accessCode string
	mux.Handle("/code", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			accessCode = r.URL.Query()["code"][0]
			fmt.Fprintln(w, "Success")
			cancel()
		},
	))

	server := &http.Server{
		Addr:    ip + ":" + serverPort,
		Handler: mux,
	}

	log.Println("Starting server on port " + serverPort)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	<-ctx.Done()
	server.Shutdown(ctx)
	log.Println("Code received, server shutdown.")

	return accessCode, redirectURI
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
