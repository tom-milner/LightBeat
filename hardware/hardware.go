package hardware

import (
	"github.com/ikester/blinkt"
)

func SetupLights() {
	bl := blinkt.NewBlinkt()
	bl.Setup()
	bl.ShowInitialAnim()
}
