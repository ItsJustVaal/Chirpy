package main

import (
	"log"
	"net/http"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := "42069"
	server := http.NewServeMux()
	server.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	server.HandleFunc("/healthz", handlerReadiness)

	corServer := middlewareCors(server)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corServer,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
