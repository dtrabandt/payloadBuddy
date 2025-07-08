package main

import (
	"encoding/json"
	"net/http"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	data := make([]Item, 100000)
	for i := 0; i < 100000; i++ {
		data[i] = Item{ID: i, Name: "Object"}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
