package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type DB struct {
	path        string
	chirpsCount int
	chirps      DBChirp
	mux         *sync.RWMutex
}

type DBChirp struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Body string `json:"body"`
	ID   int    `json:"id"`
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
		err := os.WriteFile("./database.json", []byte("{ \"chirps\": {}}"), 0666)
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
	}
	err = DB.loadDB()
	if err != nil {
		return &DB, err
	}
	return &DB, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	if body == "" {
		return Chirp{}, fmt.Errorf("Body is empty")
	}
	newChirp := Chirp{
		ID:   db.chirpsCount,
		Body: body,
	}
	db.chirps.Chirps[db.chirpsCount] = newChirp
	db.chirpsCount++
	err := db.writeDB()
	if err != nil {
		log.Fatalln(err)
	}
	return newChirp, nil
}

func (db *DB) GetChirps() (DBChirp, error) {
	return db.chirps, nil
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
	data, err := json.Marshal(db.chirps.Chirps)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.WriteFile("./database.json", data, 0666)
	if err != nil {
		return fmt.Errorf("Error Writing to DB, %s", err)
	}
	log.Println("Database saved")
	return nil
}
