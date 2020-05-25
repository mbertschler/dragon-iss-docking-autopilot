package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var rotateParameters = ControllerParameters{
	DampingCycles: 2.0,
	Correction:    0.5,
	RateFactor:    0.5,
	RateMin:       0.1,
	RateMax:       1.0,
}

var approachParameters = ControllerParameters{
	DampingCycles: 4.0,
	Correction:    0.3,
	RateFactor:    .1,
	RateMin:       0.15,
	RateMax:       10.0,
}

var centerParameters = ControllerParameters{
	DampingCycles: 4.0,
	Correction:    0.3,
	RateFactor:    0.2,
	RateMin:       0.05,
	RateMax:       5.0,
}

var IOs = []*ControlledIO{
	{
		InputSelector:          "#roll .error",
		OutputPositiveSelector: "#roll-left-button",
		OutputNegativeSelector: "#roll-right-button",
		Controller: Controller{
			ControllerParameters: rotateParameters,
		},
	},
	{
		InputSelector:          "#pitch .error",
		OutputPositiveSelector: "#pitch-up-button",
		OutputNegativeSelector: "#pitch-down-button",
		Controller: Controller{
			ControllerParameters: rotateParameters,
		},
	},
	{
		InputSelector:          "#yaw .error",
		OutputPositiveSelector: "#yaw-left-button",
		OutputNegativeSelector: "#yaw-right-button",
		Controller: Controller{
			ControllerParameters: rotateParameters,
		},
	},

	{
		InputSelector:          "#x-range > div",
		OutputPositiveSelector: "#translate-backward-button",
		OutputNegativeSelector: "#translate-forward-button",
		Controller: Controller{
			ControllerParameters: approachParameters,
		},
	},
	{
		InputSelector:          "#y-range > div",
		OutputPositiveSelector: "#translate-right-button",
		OutputNegativeSelector: "#translate-left-button",
		Controller: Controller{
			ControllerParameters: centerParameters,
		},
	},
	{
		InputSelector:          "#z-range > div",
		OutputPositiveSelector: "#translate-up-button",
		OutputNegativeSelector: "#translate-down-button",
		Controller: Controller{
			ControllerParameters: centerParameters,
		},
	},
}

func main() {
	IOs[1].Print = true
	for {
		time.Sleep(100 * time.Millisecond)
		now := time.Now()
		for _, io := range IOs {
			io.Control(now)
		}
	}
}

type ControlledIO struct {
	Controller
	InputSelector          string
	OutputPositiveSelector string
	OutputNegativeSelector string
}

func (c *ControlledIO) Control(now time.Time) {
	input := readNumber(c.InputSelector)
	clicks := c.Controller.Correct(now, input)
	button := c.OutputPositiveSelector
	if clicks < 0 {
		clicks *= -1
		button = c.OutputNegativeSelector
	}
	for i := 0; i < clicks; i++ {
		clickButton(button)
	}
}

type ControllerParameters struct {
	Correction    float64
	DampingCycles float64
	RateFactor    float64
	RateMin       float64
	RateMax       float64
	Print         bool
}

type Controller struct {
	ControllerParameters

	previousTime   time.Time
	previousOffset float64
	rate           float64
	clicks         float64
}

func (c *Controller) Correct(now time.Time, offset float64) int {
	if c.previousTime.IsZero() {
		c.previousOffset = offset
		c.previousTime = now
		return 0
	}

	instantRate := (offset - c.previousOffset) / now.Sub(c.previousTime).Seconds()
	c.rate = (c.rate*c.DampingCycles + instantRate) / (c.DampingCycles + 1)
	target := offset * -c.Correction
	offAbs := math.Abs(offset)
	limitRate := offAbs * c.RateFactor
	if limitRate < c.RateMin {
		limitRate = c.RateMin
	}
	if limitRate > c.RateMax {
		limitRate = c.RateMax
	}
	target = limit(target, limitRate)

	correction := target - c.rate
	c.clicks += correction
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
