package main

import (
	"context"
	"fmt"
	"log"
	"main/spotify"
	"main/spotify/models"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func main() {

	// Authenticate with spotify API.
	tokenFile := "tokens.json"
	if !spotify.Authorize(tokenFile) {
		log.Fatal("Failed to authorize spotify wrapper")
	}

	log.Println("Starting ticker")

	lastPlaying := spotify.GetCurrentlyPlaying()
	ticker := time.NewTicker(2 * time.Second)
	var beatContex context.Context
	var cancel context.CancelFunc
	isPlaying := lastPlaying.IsPlaying
	isDetecting := false

	for {
		<-ticker.C
		currPlay := spotify.GetCurrentlyPlaying()

		// If there has been a change in play state
		if currPlay.IsPlaying != isPlaying {
			log.Println("Change in play state")
			isPlaying = currPlay.IsPlaying
			if currPlay.IsPlaying {
				log.Println("Starting")
				beatContex, cancel = context.WithCancel(context.Background())
				go detectBeats(beatContex, currPlay)
				isDetecting = true
			} else {
				log.Println("Stopping")
				cancel()
				isDetecting = false
				continue
			}
		}

		playButNotDetect := (!isDetecting && currPlay.IsPlaying)
		// If there has been a change in media
		if currPlay.Item.ID != lastPlaying.Item.ID || playButNotDetect {

			if !playButNotDetect {
				log.Println("Change in media")
				cancel()
			} else {
				log.Println("Playing without detection")
			}
			beatContex, cancel = context.WithCancel(context.Background())
			go detectBeats(beatContex, currPlay)
			lastPlaying = currPlay
			isDetecting = true
			log.Println("Starting")
		}

	}
	// Check for a song every 2 seconds.

	// for {
	// 	<-ticker.C
	// 	currPlay := spotify.GetCurrentlyPlaying()
	// 	if currPlay.IsPlaying {
	// 		detectBeats(currPlay)
	// 	}
	// }

}

func detectBeats(ctx context.Context, currPlay models.CurrentlyPlaying) {
	log.Println("Trackin beats.")
	trackAn := spotify.GetTrackAnalysis(currPlay.Item.ID)

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
	fmt.Println("Trigger:", triggerNum)
}
