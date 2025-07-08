package main

import "net/http"

func main() {
	http.HandleFunc("/data", DataHandler)
	http.ListenAndServe(":8080", nil)
}
