package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var window = js.Global()

func main() {
	for {
		time.Sleep(100 * time.Millisecond)
		input := readInput()
		output := controller(input)
		writeOutput(output)
	}
}

type Controller struct {
	Correction  float64
	Damping     float64
	MaxRateFunc func(offset float64) float64
	Print       bool

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
	c.rate = (c.rate*c.Damping + instantRate) / (c.Damping + 1)
	target := offset * -c.Correction
	if c.MaxRateFunc != nil {
		offAbs := math.Abs(offset)
		target = limit(target, c.MaxRateFunc(offAbs))
	}
	correction := target - c.rate
	c.clicks += correction
	fullClicks := math.Round(c.clicks)
	c.clicks -= fullClicks

	if c.Print {
		fmt.Printf("target %+.3f rate %+.3f correction %+.3f clicks %+.3f fullClicks %+.0f\n",
			target, c.rate, correction, c.clicks, fullClicks)
	}

	c.previousOffset = offset
	c.previousTime = now
	return int(fullClicks)
}

type Vector struct {
	X, Y, Z float64
}

type Input struct {
	Range       Vector
	Orientation Vector
	Rotation    Vector
	Distance    float64
	Rate        float64
	Time        time.Time
}

func readInput() *Input {
	i := &Input{}
	i.Range.X = selectorValue("#x-range > div")
	i.Range.Y = selectorValue("#y-range > div")
	i.Range.Z = selectorValue("#z-range > div")
	i.Orientation.X = selectorValue("#roll .error")
	i.Orientation.Y = selectorValue("#pitch .error")
	i.Orientation.Z = selectorValue("#yaw .error")
	i.Rotation.X = selectorValue("#roll .rate")
	i.Rotation.Y = selectorValue("#pitch .rate")
	i.Rotation.Z = selectorValue("#yaw .rate")
	i.Distance = selectorValue("#range .rate")
	i.Rate = selectorValue("#rate .rate")
	i.Time = time.Now()
	return i
}

type Output struct {
	Commands []string
}

func (o *Output) Add(operation string, times int) {
	for i := 0; i < times; i++ {
		o.Commands = append(o.Commands, operation)
	}
}

func maxRateX(offset float64) float64 {
	min := .15
	max := 10.0
	rate := offset * .1
	if rate < min {
		return min
	}
	if rate > max {
		return max
	}
	return rate
}

func maxRateYZ(offset float64) float64 {
	min := .05
	rate := offset * .2
	if rate < min {
		return min
	}
	return rate
}

func maxRateRot(offset float64) float64 {
	min := 1.0
	rate := offset * 1.5
	if rate < min {
		return min
	}
	return rate
}

var (
	RotateXController = Controller{Damping: 0, Correction: .4, MaxRateFunc: maxRateRot}
	RotateYController = Controller{Damping: 0, Correction: .4, MaxRateFunc: maxRateRot}
	RotateZController = Controller{Damping: 0, Correction: .4, MaxRateFunc: maxRateRot}

	TranslateXController = Controller{Damping: 4, Correction: .3, MaxRateFunc: maxRateX}
	TranslateYController = Controller{Damping: 4, Correction: .3, MaxRateFunc: maxRateYZ}
	TranslateZController = Controller{Damping: 4, Correction: .3, MaxRateFunc: maxRateYZ}
)

func controller(in *Input) *Output {
	out := &Output{}
	now := time.Now()

	applyCorrection(RotateXController.Correct(now, in.Orientation.X), out, RotateXNeg, RotateXPos)
	applyCorrection(RotateYController.Correct(now, in.Orientation.Y), out, RotateYNeg, RotateYPos)
	applyCorrection(RotateZController.Correct(now, in.Orientation.Z), out, RotateZNeg, RotateZPos)

	applyCorrection(TranslateXController.Correct(now, in.Range.X), out, TranslateXPos, TranslateXNeg)
	applyCorrection(TranslateYController.Correct(now, in.Range.Y), out, TranslateYPos, TranslateYNeg)
	applyCorrection(TranslateZController.Correct(now, in.Range.Z), out, TranslateZPos, TranslateZNeg)

	return out
}

func applyCorrection(clicks int, out *Output, pos, neg string) {
	if clicks > 0 {
		out.Add(pos, clicks)
	} else {
		out.Add(neg, clicks*-1)
	}
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

const (
	RotateXNeg    = "#roll-left-button"
	RotateXPos    = "#roll-right-button"
	RotateYNeg    = "#pitch-up-button"
	RotateYPos    = "#pitch-down-button"
	RotateZNeg    = "#yaw-left-button"
	RotateZPos    = "#yaw-right-button"
	TranslateXPos = "#translate-backward-button"
	TranslateXNeg = "#translate-forward-button"
	TranslateYPos = "#translate-right-button"
	TranslateYNeg = "#translate-left-button"
	TranslateZPos = "#translate-up-button"
	TranslateZNeg = "#translate-down-button"
)

func writeOutput(out *Output) {
	for _, cmd := range out.Commands {
		window.Call("$", cmd).Call("click")
	}
}

func selectorValue(selector string) float64 {
	text := window.Call("$", selector).Get("innerText")
	num := extractNumber(text.String())
	out, err := strconv.ParseFloat(num, 64)
	if err != nil && num != "" {
		log.Println("ParseFloat ERROR:", err)
	}
	return out
}

func extractNumber(in string) string {
	parts := strings.Split(in, " ")
	return strings.TrimSuffix(parts[0], "°")
}
