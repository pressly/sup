package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	config, err := os.ReadFile("/etc/example.cfg")
	if err != nil {
		log.Fatal(err)
	}

	// Define handler that returns "Hello $ENV"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello "))
		if err != nil {
			panic(err)
		}

		_, err = w.Write(config)
		if err != nil {
			panic(err)
		}
	})

	server := &http.Server{
		Addr:              ":8000",
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
