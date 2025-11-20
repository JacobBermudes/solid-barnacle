package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Api_req struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Method   string `json:"method"`
	Type     string `json:"type"`
	Props    string `json:"props"`
}

type Api_resp struct {
	Username   string   `json:"username"`
	Balance    int64    `json:"balance"`
	Tariff     string   `json:"tariff"`
	SharedKeys []string `json:"sharedkey"`
	Active     string   `json:"active"`
}

func main() {

	r := http.NewServeMux()
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("API supports only POST method!"))
			return
		}

		var req Api_req
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Error parsing BODY "+err.Error(), http.StatusBadRequest)
			return
		}

		var msgs []string

		if req.Type == "wa" {
			msgs = append(msgs, "WebApp API is under construction...", fmt.Sprint(req))
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(msgs)
	})

	log.Println("Go API listening :8000 (HTTP)")

	// if err := http.ListenAndServe("127.0.0.1:8000", r); err != nil {
	// 	log.Fatal("HTTP WebHook-Server FAULT:", err)
	// }
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal("HTTP Web-Server for API FAULT:", err)
	}
}
