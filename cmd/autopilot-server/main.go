package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

const port = "8000"

func main() {
	cmd := exec.Command("go", "build", "-o", "build/autopilot.wasm")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Serving " + dir + " at http://localhost:" + port)
	fileServer := http.FileServer(http.Dir(dir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		fileServer.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
