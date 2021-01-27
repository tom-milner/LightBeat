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

func ShowLightAnimation() {
	bl.ShowInitialAnim()
}

func FlashLights() {
	bl.FlashAll(1, utils.GenRandomHexCode())
}

func FlashSequence(color string, fullTime time.Duration, forward bool) {
	numPixels := 8
	pixelDuration := time.Duration(int(fullTime) / 8)
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
	mainPixelBrightness := 0.5
	neighbourPixelBrightness := 0.04
	for c := 0; c < numPixels; c++ {
		bl.SetBrightness(0)
		bl.SetPixelBrightness(i, mainPixelBrightness)
		if i == 0 {
			bl.SetPixelBrightness(1, neighbourPixelBrightness)
		} else if i == 7 {
			bl.SetPixelBrightness(6, neighbourPixelBrightness)
		} else {
			bl.SetPixelBrightness(i+1, neighbourPixelBrightness)
			bl.SetPixelBrightness(i-1, neighbourPixelBrightness)
		}
		bl.Show()
		time.Sleep(pixelDuration)

		i += incr
	}
}
