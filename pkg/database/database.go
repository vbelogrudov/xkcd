package database

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/exp/maps"
)

type Comics struct {
	ID       int      `json:"-"`
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

type DB struct {
	path string
	data map[int]Comics
}

func New(path string) *DB {
	db := DB{
		path: path,
		data: make(map[int]Comics),
	}
	if err := db.load(); err != nil {
		log.Printf("could not load existing db: %v", err)
	}
	return &db
}

func (db *DB) Add(c Comics) {
	db.data[c.ID] = c
}

func (db *DB) Keys() []int {
	return maps.Keys(db.data)
}

func (db *DB) Size() int {
	return len(db.data)
}

func (db *DB) Save() error {
	log.Println("saving db to file")
	f, err := os.Create(db.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(db.data)
}

func (db *DB) load() error {
	f, err := os.Open(db.path)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(&db.data)
}
