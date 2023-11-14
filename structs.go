package main

type apiConfig struct {
	fileserverHits int
	database       *DB
}

type jsonBody struct {
	Body  string `json:"body"`
	Email string `json:"email"`
}

type cleanedBody struct {
	Resp string `json:"cleaned_body"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type chirpsResponse struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
}

type userResponse struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
}
