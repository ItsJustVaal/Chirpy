package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Create and load DB
	dbPath := "database.json"
	database, err := NewDB(dbPath)
	if err != nil {
		log.Fatalln(err)
	}

	cfg := apiConfig{
		fileserverHits: 0,
		database:       database,
	}

	port := "42069"
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	// Main router
	server := chi.NewRouter()
	server.Handle("/app", cfg.middlewareMetricsInc(handler))
	server.Handle("/app/*", cfg.middlewareMetricsInc(handler))

	// API sub-router
	apiServer := chi.NewRouter()
	apiServer.Get("/healthz", handlerReadiness)
	apiServer.Get("/reset", cfg.handlerReset)
	apiServer.Post("/chirps", cfg.handlerValidateChirp)
	apiServer.Get("/chirps", cfg.handlerGetChirps)

	// Admin sub-router
	adminServer := chi.NewRouter()
	adminServer.Get("/metrics", cfg.handlerMetrics)

	// Mounting sub-routers
	server.Mount("/api", apiServer)
	server.Mount("/admin", adminServer)

	// Setting CORS & Server Struct
	corServer := middlewareCors(server)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corServer,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
