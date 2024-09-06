package main

import (
	"log"
	"net/http"
	"path"
)

func main() {
	// Custom handler for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, path.Join("..", "front-end", "HTML", "index.html"))
			return
		}

		// Otherwise, use the default file server
		http.FileServer(http.Dir("../front-end")).ServeHTTP(w, r)
	})

	// Start the HTTP server on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}
