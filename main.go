package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Create and load DB
	godotenv.Load()
	dbPath := "database.json"
	database, err := NewDB(dbPath)
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	if err != nil {
		log.Fatalln(err)
	}

	cfg := apiConfig{
		fileserverHits: 0,
		JWTSecret:      jwtSecret,
		PolkaKey:       polkaKey,
		database:       database,
	}

	port := "42069"
	handler := cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	// Main router
	server := chi.NewRouter()
	server.Handle("/app", handler)
	server.Handle("/app/*", handler)

	// API sub-router
	apiServer := chi.NewRouter()
	apiServer.Get("/healthz", handlerReadiness)
	apiServer.Get("/reset", cfg.handlerReset)
	apiServer.Get("/chirps", cfg.handlerGetChirps)
	apiServer.Get("/chirps/{id}", cfg.handlerGetChirpByID)

	apiServer.Post("/chirps", cfg.handlerValidateChirp)
	apiServer.Post("/users", cfg.handlerAddUser)
	apiServer.Post("/login", cfg.handlerLogin)
	apiServer.Post("/refresh", cfg.handlerRefresh)
	apiServer.Post("/revoke", cfg.handlerRevoke)
	apiServer.Post("/polka/webhooks", cfg.handlerUpgradeUser)

	apiServer.Put("/users", cfg.handlerUpdateUser)

	apiServer.Delete("/chirps/{id}", cfg.handlerDeleteChirpByID)

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
