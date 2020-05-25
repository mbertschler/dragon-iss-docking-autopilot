package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var rotateParameters = controllerParameters{
	DampingCycles: 2.0,
	Correction:    0.5,
	RateFactor:    0.5,
	RateMin:       0.2,
	RateMax:       1.0,
}

// Y/Z translation to center on the docking port
var centerParameters = controllerParameters{
	DampingCycles: 8.0,
	Correction:    0.3,
	RateFactor:    0.2,
	RateMin:       0.1,
	RateMax:       1.0,
}

// X translation towards the station
var approachParameters = controllerParameters{
	DampingCycles: 8.0,
	Correction:    0.2,
	RateFactor:    .1,
	RateMin:       0.15,
	RateMax:       3.0,
}

var ios = []*controlledIO{
	{
		InputSelector:          "#roll .error",
		OutputPositiveSelector: "#roll-left-button",
		OutputNegativeSelector: "#roll-right-button",
		Controller: controller{
			controllerParameters: rotateParameters,
		},
	},
	{
		InputSelector:          "#pitch .error",
		OutputPositiveSelector: "#pitch-up-button",
		OutputNegativeSelector: "#pitch-down-button",
		Controller: controller{
			controllerParameters: rotateParameters,
		},
	},
	{
		InputSelector:          "#yaw .error",
		OutputPositiveSelector: "#yaw-left-button",
		OutputNegativeSelector: "#yaw-right-button",
		Controller: controller{
			controllerParameters: rotateParameters,
		},
	},

	{
		InputSelector:          "#x-range > div",
		OutputPositiveSelector: "#translate-backward-button",
		OutputNegativeSelector: "#translate-forward-button",
		Controller: controller{
			controllerParameters: approachParameters,
		},
	},
	{
		InputSelector:          "#y-range > div",
		OutputPositiveSelector: "#translate-right-button",
		OutputNegativeSelector: "#translate-left-button",
		Controller: controller{
			controllerParameters: centerParameters,
		},
	},
	{
		InputSelector:          "#z-range > div",
		OutputPositiveSelector: "#translate-up-button",
		OutputNegativeSelector: "#translate-down-button",
		Controller: controller{
			controllerParameters: centerParameters,
		},
	},
}

func main() {
	// main control loop at 10Hz
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		<-ticker.C
		now := time.Now()
		for _, io := range ios {
			io.Control(now)
		}
	}
}

type controlledIO struct {
	Controller             controller
	InputSelector          string
	OutputPositiveSelector string
	OutputNegativeSelector string
}

func (c *controlledIO) Control(now time.Time) {
	input := readNumber(c.InputSelector)
	clicks := c.Controller.correct(now, input)
	button := c.OutputPositiveSelector
	if clicks < 0 {
		clicks *= -1
		button = c.OutputNegativeSelector
	}
	for i := 0; i < clicks; i++ {
		clickButton(button)
	}
}

type controllerParameters struct {
	Correction    float64
	DampingCycles float64
	RateFactor    float64
	RateMin       float64
	RateMax       float64
	Print         bool
}

type controller struct {
	controllerParameters

	previousTime   time.Time
	previousOffset float64
	rate           float64
	clicks         float64
}

func (c *controller) correct(now time.Time, offset float64) int {
	if c.previousTime.IsZero() {
		// initialize if this is the first cycle
		c.previousOffset = offset
		c.previousTime = now
		return 0
	}

	// calculate the instant rate between the last cycle and now
	instantRate := (offset - c.previousOffset) / now.Sub(c.previousTime).Seconds()

	// dampen the instant rate with the previous rate according to DampingCycles
	c.rate = (c.rate*c.DampingCycles + instantRate) / (c.DampingCycles + 1)

	// proportional correction based on the offset
	target := offset * -c.Correction

	// calculate the limit rate based on RateFactor
	// while making sure rate is > RateMin and < RateMax
	limitRate := math.Abs(offset) * c.RateFactor
	if limitRate < c.RateMin {
		limitRate = c.RateMin
	}
	if limitRate > c.RateMax {
		limitRate = c.RateMax
	}

	// limit the target rate with the limit rate
	target = limit(target, limitRate)

	// difference between target and current rate
	correction := target - c.rate

	// apply correction to clicks accumulator
	c.clicks += correction

	// take full clicks from accumulator as output
	fullClicks := math.Floor(c.clicks)
	c.clicks -= fullClicks

	if c.Print {
		fmt.Printf("target %+.3f rate %+.3f correction %+.3f clicks %+.3f fullClicks %+.0f\n",
			target, c.rate, correction, c.clicks, fullClicks)
	}

	c.previousOffset = offset
	c.previousTime = now
	return int(fullClicks)
}

// limits the input regardless of sign
func limit(in, limit float64) float64 {
	if in > limit {
		in = limit
	}
	if in < -limit {
		in = -limit
	}
	return in
}

func readNumber(selector string) float64 {
	text := js.Global().Call("$", selector).Get("innerText").String()
	if text == "" {
		return 0
	}
	parts := strings.Fields(text)
	text = strings.TrimSuffix(parts[0], "°")
	num, err := strconv.ParseFloat(text, 64)
	if err != nil {
		fmt.Printf("can't parse %q as float: %v", text, err)
	}
	return num
}

func clickButton(selector string) {
	js.Global().Call("$", selector).Call("click")
}
