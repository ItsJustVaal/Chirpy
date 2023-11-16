package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func cleanInput(body string, blockedWords map[string]struct{}) string {
	splitString := strings.Split(body, " ")
	for i, x := range splitString {
		lower := strings.ToLower(x)
		if _, ok := blockedWords[lower]; ok {
			splitString[i] = "****"
		}
	}
	return strings.Join(splitString, " ")
}

// Reusable http response functions
// Errorfunc
func errorResp(w http.ResponseWriter, errorCode int, message string) {
	log.Printf("Responding error: %s", message)
	jsonResp(w, errorCode, errorResponse{
		Error: message,
	})
}

// Standard response
func jsonResp(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "applcation/json")
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marhsalling json: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

// Hash password
func hashPassword(password string) (string, error) {
	newPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return " ", err
	}
	return string(newPass), nil

}

func sortChirps(chirps []chirpsResponse, sortOrder string) []chirpsResponse {
	log.Println("Inside sort")
	log.Printf("Sort Method, %s", sortOrder)
	if sortOrder == "asc" || sortOrder == "" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	}

	return chirps
}
