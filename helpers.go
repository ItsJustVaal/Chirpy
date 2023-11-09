package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func errorResp(w http.ResponseWriter, errorCode int, message string) {
	if errorCode > 499 {
		log.Printf("Responding with 5xx error: %s", message)
	}
	jsonResp(w, errorCode, errorResponse{
		Error: message,
	})
}

func jsonResp(w http.ResponseWriter, statusCode int, response interface{}) {
	log.Println(response)
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