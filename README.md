# Dragon ISS Docking Autopilot in Go

Autopilot written in Go and executed as WebAssembly for docking the SpaceX
Dragon capsule in the [official simulator](https://iss-sim.spacex.com). 

### [Screencast of the autopilot in action](https://youtu.be/jLTr6UwuSd4)

## Why

I never used Go and WebAssembly together before. This seemed like a nice and
small excercise and I also got to refresh controller algorithms.

## Using this autopilot

- Compile the autopilot and serve the files.

```bash
go run ./cmd/autopilot-server
```

- Open the SpaceX ISS docking simulator: https://iss-sim.spacex.com
- Open the developer console of your browser (eg. right click > Inspect in Chrome)
- Paste the following code into the console to load the autopilot:

```js
const s = document.createElement("script");
s.setAttribute("src", "http://localhost:8000/loader.js?t="+ new Date().getTime());
document.body.appendChild(s);
```

If you want to make modifications to `autopilot.go`, restart the
`autopilot-server`, refresh the simulator page and paste the above
JavaScript code into the console again.

## Controller algorithm

The main controller algorithm is the `correct()` method in `autopilot.go`.
At first the current rate is calculated using the previous time and offset
and dampened over some cycles `DampingCycles`.

Then a correction factor `Correction` is used to calculate the target rate.
This target rate is limited based on the offset and `RateFactor` but kept
between `RateMin` and `RateMax`.

The difference between target and current rate is then added to a clicks
accumulator. Full clicks are then subtracted from the accumulator and
returned from the function.

Play around with the `ios` configuration to get different controller behavior.

---

Built by [@mbertschler](https://twitter.com/mbertschler)
