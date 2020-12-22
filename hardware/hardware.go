package hardware

import (
	"github.com/ikester/blinkt"
)

var bl blinkt.Blinkt

func SetupLights() {
	bl := blinkt.NewBlinkt()
	bl.Setup()
	bl.ShowInitialAnim()
}

func FlashLights() {
	bl.FlashAll(1, "FFCC66")
}
