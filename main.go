package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/iot"
	"main/iot/topics"
	"main/spotify"
	"main/spotify/models"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func main() { // Setup

	// Authenticate with spotify API.
	tokenFile := "tokens.json"
	if !spotify.Authorize(tokenFile) {
		log.Fatal("Failed to authorize spotify wrapper")
	}

	// Connect to MQTT broker
	broker := iot.MQTTBroker{
		Address: "localhost",
		Port:    "1883",
	}
	info := iot.MQTTConnInfo{
		ClientID: "LightBeatGateway",
		Broker:   broker,
	}

	_, err := iot.ConnectToMQTTBroker(info)
	if err != nil {
		log.Fatal(err)
	}

	run()
}

func run() {
	log.Println("Starting ticker")
	lastPlaying := spotify.GetCurrentlyPlaying()
	ticker := time.NewTicker(2 * time.Second)
	var beatContex context.Context
	var cancel context.CancelFunc
	isDetecting := false

	for {
		<-ticker.C
		currPlay := spotify.GetCurrentlyPlaying()

		changeInPlayState := lastPlaying.IsPlaying != currPlay.IsPlaying
		changeInMedia := lastPlaying.Item.ID != currPlay.Item.ID
		playingWithoutDetection := (!isDetecting && currPlay.IsPlaying)

		if ((changeInPlayState && !currPlay.IsPlaying) || changeInMedia) && !playingWithoutDetection {
			log.Println("Stopping")
			cancel()
			isDetecting = false
		}

		if ((changeInPlayState && currPlay.IsPlaying) || changeInMedia) || playingWithoutDetection {
			log.Println("Starting")
			beatContex, cancel = context.WithCancel(context.Background())
			go detectBeats(beatContex, currPlay)
			isDetecting = true
		}

		lastPlaying = currPlay
	}
}

func detectBeats(ctx context.Context, currPlay models.Media) {
	log.Println("Trackin beats.")
	trackAn := spotify.GetTrackAnalysis(currPlay.Item.ID)

	b, _ := json.Marshal(currPlay)
	go iot.SendMessage(topics.NewMedia, b)

	fmt.Println(currPlay.Item.Name)

	// Calculate when to show the first beat.
	triggers := trackAn.Beats
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
	iot.SendMessage(topics.Beat, message)
	log.Println(message)
}
