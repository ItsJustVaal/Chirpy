package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeJWTAccess(userID int, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   fmt.Sprintf("%d", userID),
	})
	return token.SignedString(signingKey)
}

func MakeJWTRefresh(userID int, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-refresh",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   fmt.Sprintf("%d", userID),
	})
	return token.SignedString(signingKey)
}

func GetBearerToken(headers http.Header) (string, error) {
	log.Println("Inside get token")
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("No auth included")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 {
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	log.Println("Inside validate access")
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}
	checker, err := token.Claims.GetIssuer()
	if checker == "chirpy-refresh" {
		return "", fmt.Errorf("Invalid token in validate. Refresh Token Found")
	}
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return userIDString, nil
}

func ValidateJWTRefresh(tokenString, tokenSecret string) (string, error) {
	log.Println("Inside validate refresh")
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}
	checker, err := token.Claims.GetIssuer()
	if checker != "chirpy-refresh" {
		return "", fmt.Errorf("Invalid token in validate. Refresh Token Not Found")
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return userIDString, err
}
