package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (cfg *apiConfig) AddChirp(body string) (Chirp, error) {
	newChirp, err := cfg.database.CreateChirp(body)
	if err != nil {
		log.Fatalln("Failed to add Chirp")
		return Chirp{}, err
	}
	return newChirp, nil
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	newUser, err := cfg.database.CreateUser(checker.Email, checker.Password)
	if err != nil {
		log.Fatalln("Failed to add User")
		return
	}
	jsonResp(w, http.StatusCreated, userResponse{
		Email: newUser.Email,
		ID:    newUser.ID,
	})
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	if len(checker.Body) > 140 {
		errorResp(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	blockedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleanBody := cleanInput(checker.Body, blockedWords)
	newChirp, err := cfg.AddChirp(cleanBody)
	if err != nil {
		log.Fatalln(err)
	}
	jsonResp(w, http.StatusCreated, chirpsResponse{
		Body: newChirp.Body,
		ID:   newChirp.ID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.database.GetChirps()
	if err != nil {
		errorResp(w, http.StatusNoContent, "No Chirps")
		return
	}

	var finalChirps []chirpsResponse
	for _, y := range allChirps.Chirps {
		finalChirps = append(finalChirps, chirpsResponse{
			Body: y.Body,
			ID:   y.ID,
		})
	}
	finalResp, err := json.Marshal(finalChirps)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Failed to Marshal")
		return
	}
	w.Write(finalResp)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := chi.URLParam(r, "id")
	log.Println(chirpIDString)
	id, err := strconv.Atoi(chirpIDString)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Error getting ID")
		return
	}
	chirp, err := cfg.database.GetChirpByID(id)
	if err != nil {
		errorResp(w, http.StatusNotFound, "Chirp Doesn't Exist")
		return
	}
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Failed to Marshal")
		return
	}

	jsonResp(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	id, err := cfg.database.checkLogin(checker.Email, checker.Password)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Wrong email or password")
	}
	jsonResp(w, http.StatusOK, userResponse{
		Email: checker.Email,
		ID:    id,
	})
}
