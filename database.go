package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type DB struct {
	path        string
	chirpsCount int
	usersCount  int
	chirps      DBChirp
	mux         *sync.RWMutex
}

type DBChirp struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RevokedTokens map[string]time.Time `json:"revoked_tokens"`
}

type Chirp struct {
	Author int    `json:"author_id"`
	Body   string `json:"body"`
	ID     int    `json:"id"`
}

type User struct {
	Password  string `json:"password"`
	Email     string `json:"email"`
	ID        int    `json:"id"`
	ChirpyRed bool   `json:"is_chirpy_red"`
}

func NewDB(path string) (*DB, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			os.Create("database.json")
			log.Println("Database file created")
		}
	} else {
		log.Println("Database file found, deleting...")
		os.Remove("database.json")
		log.Println("Deleted old file")
		log.Println("Creating new file")
		os.Create("database.json")
		err := os.WriteFile("./database.json", []byte("{ \"chirps\": {}, \"users\": {}, \"revoked_tokens\": {}}"), 0666)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if path == "" {
		return &DB{}, fmt.Errorf("Path is empty")
	}
	DB := DB{
		path:        path,
		chirpsCount: 1,
		usersCount:  1,
	}
	err = DB.loadDB()
	if err != nil {
		return &DB, err
	}
	return &DB, nil
}

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	if body == "" {
		return Chirp{}, fmt.Errorf("Body is empty")
	}
	newChirp := Chirp{
		ID:     db.chirpsCount,
		Body:   body,
		Author: authorID,
	}
	db.chirps.Chirps[db.chirpsCount] = newChirp
	db.chirpsCount++
	err := db.writeDB()
	if err != nil {
		log.Fatalln(err)
	}
	return newChirp, nil
}

func (db *DB) DeleteChirp(chirpID, authorID int) error {
	checker, err := db.GetChirpByID(chirpID)
	if err != nil {
		return err
	}

	if checker.Author != authorID {
		return fmt.Errorf("Not the correct author")
	}

	delete(db.chirps.Chirps, chirpID)
	return nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	if email == "" {
		return User{}, fmt.Errorf("Body is empty")
	}
	for x := range db.chirps.Users {
		if db.chirps.Users[x].Email == email {
			return User{}, fmt.Errorf("Email already exists")
		}
	}
	hashedPW, err := hashPassword(password)
	if err != nil {
		log.Fatalln(err)
	}
	newUser := User{
		Password:  hashedPW,
		ID:        db.usersCount,
		Email:     email,
		ChirpyRed: false,
	}
	db.chirps.Users[db.usersCount] = newUser
	db.usersCount++
	err = db.writeDB()
	if err != nil {
		log.Fatalln(err)
	}
	return newUser, nil
}

func (db *DB) UpdateUser(email, password string, id int) (User, error) {
	user, ok := db.chirps.Users[id]
	if !ok {
		return User{}, fmt.Errorf("User not found")
	}
	hashedPass, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}

	user.Email = email
	user.Password = hashedPass
	db.chirps.Users[id] = user

	err = db.writeDB()
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) checkLogin(email string) (User, error) {
	for _, user := range db.chirps.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, fmt.Errorf("User doesn't exist")
}

func (db *DB) checkRevokedDB(token string) error {
	_, ok := db.chirps.RevokedTokens[token]
	if !ok {
		return fmt.Errorf("Token not found")
	}
	return nil
}

func (db *DB) revokeToken(token string) error {
	_, ok := db.chirps.RevokedTokens[token]
	if !ok {
		db.chirps.RevokedTokens[token] = time.Now()
		return nil
	}
	return fmt.Errorf("Token already revoked")
}

func (db *DB) GetChirps() (map[int]Chirp, error) {
	return db.chirps.Chirps, nil
}

func (db *DB) GetChirpByID(id int) (Chirp, error) {
	chirp, ok := db.chirps.Chirps[id]
	if !ok {
		return Chirp{}, fmt.Errorf("Chirp Does Not Exist")
	}
	respChirp := Chirp{
		Author: chirp.Author,
		Body:   chirp.Body,
		ID:     chirp.ID,
	}
	return respChirp, nil
}

func (db *DB) loadDB() error {
	chirps, err := os.ReadFile("./database.json")
	if err != nil {
		log.Fatalf("Error reading file, %s", err)
	}
	err = json.Unmarshal(chirps, &db.chirps)
	if err != nil {
		log.Fatalf("Error loading in chirps to memeory, %s", err)
	}
	log.Println("Chirps loaded into memory")
	return nil
}

func (db *DB) writeDB() error {
	data, err := json.Marshal(db.chirps)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile("./database.json", data, 0600)
	if err != nil {
		return fmt.Errorf("Error Writing to DB, %s", err)
	}
	log.Println("Database saved")
	return nil
}

func (db *DB) upgradeUser(userID int) error {
	user, ok := db.chirps.Users[userID]
	if !ok {
		return fmt.Errorf("User not found")
	}
	user.ChirpyRed = true
	db.chirps.Users[userID] = user

	err := db.writeDB()
	if err != nil {
		return err
	}
	return nil
}
