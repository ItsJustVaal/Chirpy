package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Checks Server Status
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// Returns total hits
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
	
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>
		`, cfg.fileserverHits)))
}

// Increments total hits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

// Resets total hits
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err != nil {
		w.WriteHeader(500)
		jsonErr := jsonBody{
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(jsonErr)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
		return
	}
	if len(checker.Body) > 140 {
		w.WriteHeader(400)
		jsonErr := jsonBody{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(jsonErr)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
		return
	}
	jsonResp := jsonBody{
		Valid: true,
	}
	dat, err := json.Marshal(jsonResp)
	w.WriteHeader(200)
	w.Write(dat)
	return
}
