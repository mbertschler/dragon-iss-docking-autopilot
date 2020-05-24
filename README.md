# Dragon ISS Docking Autopilot in Go

Autopilot written in Go and executed as WebAssembly for docking the SpaceX
Dragon capsule in the simulator. https://iss-sim.spacex.com

## Why

I never used Go and WebAssembly together before. This seemed like a nice and
small excercise and I also got to refresh my knowledge of controller algorithms.

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

---

Built by [@mbertschler](https://twitter.com/mbertschler)
