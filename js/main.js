/* copy this into the browser console to run this script:
const s = document.createElement("script");
s.setAttribute("src", "http://localhost:8000/js/main.js?t="+ new Date().getTime());
document.body.appendChild(s);
*/

const port = "8000"

const time = new Date().getTime();
const script = document.createElement("script");
script.setAttribute("src", "http://localhost:" + port + "/js/go_wasm_exec.js?t=" + time);
script.setAttribute("onload", "run()");
document.body.appendChild(script);

function run() {
    const go = new Go();
    const code = fetch("http://localhost:" + port + "/build/autopilot.wasm?t=" + time);
    WebAssembly.instantiateStreaming(code, go.importObject).then((result) => {
        go.run(result.instance);
    });
}
