package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
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

// Creates Chirp
func (cfg *apiConfig) AddChirp(body string, id int) (Chirp, error) {
	newChirp, err := cfg.database.CreateChirp(body, id)
	if err != nil {
		log.Fatalln("Failed to add Chirp")
		return Chirp{}, err
	}
	return newChirp, nil
}

// Creates user
func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Add User")
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
		Email:     newUser.Email,
		ID:        newUser.ID,
		ChirpyRed: newUser.ChirpyRed,
	})
}

// Updates user
func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Update User")
	token, err := GetBearerToken(r.Header)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	user, err := ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Couldn't Validate Token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err = decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	userIDInt, err := strconv.Atoi(user)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}
	updatedUser, err := cfg.database.UpdateUser(checker.Email, checker.Password, userIDInt)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, err.Error())
	}
	jsonResp(w, http.StatusOK, userResponse{
		Email: updatedUser.Email,
		ID:    userIDInt,
	})
}

// Checks if chirp is valid
func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := GetBearerToken(r.Header)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	user, err := ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, err.Error())
		return
	}
	id, err := strconv.Atoi(user)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Error getting ID")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err = decoder.Decode(&checker)
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
	newChirp, err := cfg.AddChirp(cleanBody, id)
	if err != nil {
		log.Fatalln(err)
	}
	jsonResp(w, http.StatusCreated, chirpsResponse{
		Author: newChirp.Author,
		Body:   newChirp.Body,
		ID:     newChirp.ID,
	})
}

// Gets all chirps
func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.database.GetChirps()
	if err != nil {
		errorResp(w, http.StatusNoContent, "No Chirps")
		return
	}
	var id int
	s := r.URL.Query().Get("author_id")
	if s != "" {
		id, err = strconv.Atoi(s)
		if err != nil {
			errorResp(w, http.StatusInternalServerError, "Error getting ID")
			return
		}
	}
	var finalChirps []chirpsResponse

	if id != 0 {
		for _, y := range allChirps {
			if y.Author == id {
				finalChirps = append(finalChirps, chirpsResponse{
					Author: y.Author,
					Body:   y.Body,
					ID:     y.ID,
				})
			}
		}
	} else {
		for _, y := range allChirps {
			finalChirps = append(finalChirps, chirpsResponse{
				Author: y.Author,
				Body:   y.Body,
				ID:     y.ID,
			})
		}
	}
	sortMethod := r.URL.Query().Get("sort")
	sortedChirps := sortChirps(finalChirps, sortMethod)
	finalResp, err := json.Marshal(sortedChirps)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Failed to Marshal")
		return
	}
	w.Write(finalResp)
}

// Gets a chirp by its ID
func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := chi.URLParam(r, "id")
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

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Delete Chirp")
	chirpIDString := chi.URLParam(r, "id")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Error getting ID")
		return
	}
	token, err := GetBearerToken(r.Header)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	user, err := ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, err.Error())
		return
	}
	userID, err := strconv.Atoi(user)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Error getting ID")
		return
	}

	log.Println("Deleting")
	err = cfg.database.DeleteChirp(chirpID, userID)
	if err != nil {
		errorResp(w, http.StatusForbidden, err.Error())
	}
	jsonResp(w, http.StatusOK, "Deleted")
}

// Checks login and grants token
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Login")
	type loginResponse struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.database.checkLogin(checker.Email)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(checker.Password))
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	token, err := MakeJWTAccess(user.ID, cfg.JWTSecret, time.Duration(60)*time.Minute)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldnt make JWT Access Token")
		return
	}

	refresh, err := MakeJWTRefresh(user.ID, cfg.JWTSecret, time.Duration(1440)*time.Hour)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldnt make JWT Access Token")
		return
	}

	jsonResp(w, http.StatusOK, loginResponse{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			ChirpyRed: user.ChirpyRed,
		},
		Token:        token,
		RefreshToken: refresh,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Refresh")
	type refreshResp struct {
		Token string `json:"token"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err == nil {
		errorResp(w, http.StatusUnauthorized, "Request cant have a body")
		return
	}

	token, err := GetBearerToken(r.Header)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	user, err := ValidateJWTRefresh(token, cfg.JWTSecret)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "test 1")
		return
	}

	err = cfg.database.checkRevokedDB(token)
	if err == nil {
		errorResp(w, http.StatusUnauthorized, "Token revoked")
		return
	}
	userIDInt, err := strconv.Atoi(user)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	newToken, err := MakeJWTAccess(userIDInt, cfg.JWTSecret, time.Duration(60)*time.Minute)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Test 2")
		return
	}

	log.Println("Passed all checks, responding")
	jsonResp(w, http.StatusOK, refreshResp{
		Token: newToken,
	})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling Revoke")
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := jsonBody{}
	err := decoder.Decode(&checker)
	if err == nil {
		errorResp(w, http.StatusUnauthorized, "Request cant have a body")
		return
	}

	token, err := GetBearerToken(r.Header)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}
	_, err = ValidateJWTRefresh(token, cfg.JWTSecret)
	if err != nil {
		errorResp(w, http.StatusUnauthorized, err.Error())
		return
	}

	err = cfg.database.revokeToken(token)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResp(w, http.StatusOK, "Token Revoked")
}

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling upgrade")
	type userEvent struct {
		Event string `json:"event"`
		Data  struct {
			User int `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := GetBearerToken(r.Header)
	if err != nil || apiKey != cfg.PolkaKey {
		errorResp(w, http.StatusUnauthorized, "Unauthorized API Key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	checker := userEvent{}
	err = decoder.Decode(&checker)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, "Request has no body")
		return
	}

	if checker.Event != "user.upgraded" {
		jsonResp(w, http.StatusOK, "not an upgrade event")
		return
	}

	err = cfg.database.upgradeUser(checker.Data.User)
	if err != nil {
		errorResp(w, http.StatusNotFound, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, "")
}
