package main

import (
	"encoding/json"
	"log"
	"net/http"
	"speed-ball/handlers"
	core "speed-ball/internal/core/data"
	"strconv"
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

		User := core.User{
			UserID:   req.Id,
			Username: req.Username,
		}

		if req.Method == "start" && !User.AccountExist() && req.Props != "" {

			User.RefBonus(100)

			makerID, _ := strconv.ParseInt(req.Props, 10, 64)
			inviteMaker := core.User{
				UserID: makerID,
			}
			inviteMaker.RefBonus(100)

			User = User.SetAccount()
		}

		var msgs []string

		if req.Type == "cb" {

			cbData := handlers.CallbackHandler{
				Data: req.Method,
				User: User,
			}

			msgs = cbData.HandleCallback()
		} else if req.Type == "cmd" {

			cmdData := handlers.CommandHandler{
				Data: req.Method,
				User: User,
			}

			msgs = cmdData.HandleCommand()
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
