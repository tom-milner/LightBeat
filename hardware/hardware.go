package hardware

import (
	"time"

	"github.com/ikester/blinkt"
	"github.com/tom-milner/LightBeatGateway/utils"
)

var bl blinkt.Blinkt

func SetupLights() {
	bl := blinkt.NewBlinkt(0.75)
	bl.Setup()
	bl.ShowInitialAnim()
}

func FlashLights() {
	bl.FlashAll(1, utils.GenRandomHexCode())
}

func FlashSequence(color string, fullTime time.Duration, forward bool) {
	numPixels := 8
	pixelDuration := time.Duration(int(fullTime  - 50 * time.Millisecond ) / 8)
	bl.SetAll(blinkt.Hex2RGB(color))
	bl.SetBrightness(0)

	var incr, i int
	if forward {
		incr = 1
		i = 0
	} else {
		incr = -1
		i = numPixels - 1
	}

	for c := 0; c < numPixels; c++ {
		bl.SetPixelBrightness(i, 1)
		bl.Show()
		time.Sleep(pixelDuration)
		bl.SetPixelBrightness(i, 0)
		bl.Show()
		i += incr
	}
}
