package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	config, err := ioutil.ReadFile("/etc/example.cfg")
	if err != nil {
		log.Fatal(err)
	}

	// Define handler that returns "Hello $ENV"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello "))
		w.Write(config)
	})

	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
