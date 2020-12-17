package main

import (
	"fmt"
	"log"
	"main/spotify"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func main() {

	// Authenticate with spotify API.
	tokenFile := "tokens.json"
	if !spotify.Authorize(tokenFile) {
		log.Fatal("Failed to authorize spotify wrapper")
	}

	// Get track data.
	currPlay := spotify.GetCurrentlyPlaying()
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
		<-ticker.C
		onTrigger(nextTrigger)
		nextTrigger++
		ticker = time.NewTicker(time.Duration(triggers[nextTrigger].Duration*1000) * time.Millisecond)
	}
}

func onTrigger(triggerNum int) {

	fmt.Println("Trigger:", triggerNum)

}
