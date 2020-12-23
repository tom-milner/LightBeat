package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tom-milner/LightBeatGateway/spotify/models"
	"github.com/tom-milner/LightBeatGateway/spotify/urls"
	"github.com/tom-milner/LightBeatGateway/utils"
)

type SpotifyAPICredentials struct {
	ClientID     string
	ClientSecret string
}

var spotifyToken models.SpotifyToken
var apiCreds SpotifyAPICredentials

// buildAPIRequest returns a request that has the spotify token stored in the params.
func buildAPIRequest(method string, url string, body io.Reader) (*http.Client, *http.Request, error) {
	client := &http.Client{
		Timeout: time.Second * 5, // 10 second timeout.
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

// This function makes the http request. Here is where we can add any global response interceptors (similar to javascript axios interceptors)
// There's probably a nicer way to add interceptors. For now, this'll do.
func makeSpotifyRequest(client *http.Client, req *http.Request) (*http.Response, error) {

	// Make the request.
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return res, err
	}

	// Check if we need to refresh the access token.
	if res.StatusCode == 401 {
		log.Println("Access code invalid. Refreshing.")
		// Refresh the access token
		spotifyToken = getAccessToken(spotifyToken.Refresh)

		// Retry the original request.
		req.Header.Set("Authorization", "Bearer "+spotifyToken.Access)
		res, err = makeSpotifyRequest(client, req)
	}

	return res, err

}

// GetMediaAudioFeatures gets the audio features of given media
func GetMediaAudioFeatures(trackID string) (models.MediaAudioFeatures, error) {

	var audioFeatures models.MediaAudioFeatures

	client, req, err := buildAPIRequest("GET", urls.MediaAudioFeatures+"/"+trackID, nil)

	if err != nil {
		log.Println(err)
		return audioFeatures, err
	}

	res, err := makeSpotifyRequest(client, req)

	// Error checks
	if err != nil {
		log.Println(err)
		return audioFeatures, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if !(res.StatusCode < 300 && res.StatusCode > 100) {
		log.Println(res.Status)
		return audioFeatures, errors.New(res.Status)
	}

	// Decode the data.
	err = json.NewDecoder(res.Body).Decode(&audioFeatures)
	return audioFeatures, err
}

// GetMediaAudioAnalysis fetches the spotify audio analysis of the supplied track.
func GetMediaAudioAnalysis(trackID string) (models.MediaAudioAnalysis, error) {

	var trackAn models.MediaAudioAnalysis

	client, req, err := buildAPIRequest("GET", urls.MediaAudioAnalysis+"/"+trackID, nil)
	if err != nil {
		log.Println(err)
		return trackAn, err
	}

	res, err := makeSpotifyRequest(client, req)

	// Error checks
	if err != nil {
		log.Println(err)
		return trackAn, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if !(res.StatusCode < 300 && res.StatusCode > 100) {
		log.Println(res.Status)
		return trackAn, errors.New(res.Status)
	}

	// Decode the data.
	err = json.NewDecoder(res.Body).Decode(&trackAn)
	return trackAn, err
}

// GetCurrentlyPlaying gets the currently-playing media from spotify.
func GetCurrentlyPlaying() (models.Media, error) {
	// TODO: better error handling!
	var currPlay models.Media

	// Make the request.
	client, req, err := buildAPIRequest("GET", urls.CurrentlyPlaying, nil)
	if err != nil {
		log.Println(err)
		return currPlay, err
	}

	res, err := makeSpotifyRequest(client, req)

	// Error checks
	if err != nil {
		log.Println(err)
		return currPlay, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if !(res.StatusCode < 300 && res.StatusCode > 100) {
		log.Println(res.Status)
		return currPlay, errors.New(res.Status)
	}

	if res.StatusCode == 204 {
		return currPlay, nil
	}

	// Decode the data.
	err = json.NewDecoder(res.Body).Decode(&currPlay)
	return currPlay, err
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
func Authorize(tokenFile string, clientID string, clientSecret string) bool {
	apiCreds.ClientID = clientID
	apiCreds.ClientSecret = clientSecret
	refreshToken, err := loadRefreshToken(tokenFile)
	if err != nil {
		log.Println(err)
		code, redirectURI := fetchAuthCode()
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
	tokens := fetchSpotifyTokens(body)
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

	tokens := fetchSpotifyTokens(body)
	return tokens
}
func fetchSpotifyTokens(body url.Values) models.SpotifyToken {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("POST", urls.NewToken, strings.NewReader(body.Encode()))
	req.SetBasicAuth(apiCreds.ClientID, apiCreds.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := makeSpotifyRequest(client, req)

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

func fetchAuthCode() (string, string) {

	// The use must visit this link to authenticate spotify for the first time.
	u, err := url.Parse(urls.Code)
	if err != nil {
		log.Fatal(err)
	}
	ip := utils.GetOutboundIP().String()
	serverPort := "8080"
	serverAddress := "http://" + ip + ":" + serverPort
	redirectURI := serverAddress + "/code"
	log.Println("Current IP: " + ip)
	q := u.Query()
	q.Set("client_id", apiCreds.ClientID)
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
