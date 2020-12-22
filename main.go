package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/hardware"
	"main/iot"
	"main/iot/topics"
	"main/spotify"
	"main/spotify/models"
	"math"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const enableHardware bool = false

func main() { // Setup

	// Authenticate with spotify API.
	tokenFile := "tokens.json"
	if !spotify.Authorize(tokenFile) {
		log.Fatal("Failed to authorize spotify wrapper")
	}

	// Connect to MQTT broker
	broker := iot.MQTTBroker{
		Address: "raspberrypi.local",
		Port:    "1883",
	}
	info := iot.MQTTConnInfo{
		ClientID: "LightBeatGateway",
		Broker:   broker,
	}

	// Flash Blinkt.
	if enableHardware {
		hardware.SetupLights()
	}
	_, err := iot.ConnectToMQTTBroker(info)
	if err != nil {
		log.Fatal(err)
	}

	run()
}

func run() {
	log.Println("Starting ticker")
	lastPlaying, _ := spotify.GetCurrentlyPlaying()
	tickerInterval := 2 * time.Second
	ticker := time.NewTicker(tickerInterval)
	var beatContex context.Context
	var cancel context.CancelFunc
	isDetecting := false

	for {
		<-ticker.C
		currPlay, _ := spotify.GetCurrentlyPlaying()
		if currPlay.Item.ID == "" {
			continue
		}

		// Whether the media has stopped or started playing.
		changeInPlayState := lastPlaying.IsPlaying != currPlay.IsPlaying

		// Whether the playing media has changed.
		changeInMedia := lastPlaying.Item.ID != currPlay.Item.ID

		// Whether the progress of the media has been changed by more than it should've in the given time interval
		progressChanged := math.Abs(float64(currPlay.Progress-lastPlaying.Progress)) > float64((tickerInterval/time.Millisecond)+time.Second) // +1 second *just to be sure*L

		// Whether media is playing, but we aren't running the beat detector.
		playingWithoutDetection := (!isDetecting && currPlay.IsPlaying)

		if ((changeInPlayState && !currPlay.IsPlaying) || changeInMedia || progressChanged) && !playingWithoutDetection {
			log.Println("Stopping")
			cancel()
			isDetecting = false
		}

		if ((changeInPlayState && currPlay.IsPlaying) || changeInMedia || progressChanged) || playingWithoutDetection {
			log.Println("Starting")
			beatContex, cancel = context.WithCancel(context.Background())

			mediaAnalysis, err := spotify.GetMediaAudioAnalysis(currPlay.Item.ID)
			if err != nil {
				continue
			}
			b, _ := json.Marshal(currPlay)
			go iot.SendMessage(topics.NewMedia, b)

			mediaFeatures, err := spotify.GetMediaAudioFeatures(currPlay.Item.ID)
			if err != nil {
				continue
			}
			b, _ = json.Marshal(mediaFeatures)
			go iot.SendMessage(topics.MediaFeatures, b)

			go triggerBeats(beatContex, currPlay, mediaAnalysis)
			isDetecting = true
		}

		lastPlaying = currPlay
	}
}

func triggerBeats(ctx context.Context, currPlay models.Media, mediaAnalysis models.MediaAudioAnalysis) {
	log.Println("Tracking beats.")

	fmt.Println(currPlay.Item.Name)

	// Calculate when to show the first beat.
	triggers := mediaAnalysis.Beats
	spew.Dump(triggers[0])

	numTriggers := len(triggers)
	var nextTrigger int = 0
	progress := time.Duration(currPlay.Progress) * time.Millisecond

	fmt.Printf("Progress: %v\n", progress)

	for i := 0; i < numTriggers; i++ {
		// Find the next beat.
		if progress >= time.Duration(triggers[i].Start*1000)*time.Millisecond {
			nextTrigger = i
		}
	}

	fmt.Printf("Trigger: %v\n", nextTrigger)
	fmt.Printf("numTriggers: %v\n", numTriggers)

	ticker := time.NewTicker(time.Duration(triggers[nextTrigger].Duration*1000) * time.Millisecond)
	for nextTrigger < numTriggers-1 {
		select {
		case <-ticker.C:
			onTrigger(nextTrigger)
			nextTrigger++
			ticker = time.NewTicker(time.Duration(triggers[nextTrigger].Duration*1000) * time.Millisecond)
		case <-ctx.Done():
			log.Println("Heard cancel. Exiting")
			return
		}
	}
}

// Function to run on every beat.
func onTrigger(triggerNum int) {
	message := fmt.Sprintf("Trigger: %d", triggerNum)
	go iot.SendMessage(topics.Beat, message)
	go hardware.FlashLights()
	log.Println(message)
}
