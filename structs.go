package main

type apiConfig struct {
	fileserverHits int
}

type jsonBody struct {
	Body  string `json:"body"`
	Error string `json:"error"`
	Valid bool   `json:"valid"`
}
