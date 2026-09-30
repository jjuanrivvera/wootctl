// Command demoserver supplies invented API responses for an account-free VHS demo.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := "127.0.0.1:8646"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           http.HandlerFunc(serveDemo),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.Println("invented demo API listening")
	log.Fatal(server.ListenAndServe())
}

func serveDemo(w http.ResponseWriter, r *http.Request) {
	var result any
	switch r.Method + " " + r.URL.Path {
	case "GET /api/v1/accounts/999/conversations":
		result = map[string]any{"data": map[string]any{"meta": map[string]any{"all_count": 2}, "payload": []any{
			map[string]any{"id": 9001, "inbox_id": 901, "status": "open", "priority": "high"},
			map[string]any{"id": 9002, "inbox_id": 902, "status": "open", "priority": "medium"},
		}}}
	case "GET /api/v1/accounts/999/conversations/9001/messages":
		result = map[string]any{"payload": []any{
			map[string]any{"id": 9101, "conversation_id": 9001, "content": "How do I export my demo report?", "message_type": 0, "created_at": 1790762400, "private": false},
			map[string]any{"id": 9102, "conversation_id": 9001, "content": "Checking the export options now.", "message_type": 1, "created_at": 1790762450, "private": false},
		}}
	case "POST /api/v1/accounts/999/conversations/9001/messages":
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		result = map[string]any{"id": 9103, "conversation_id": 9001, "content": body.Content, "message_type": 1, "created_at": 1790762500, "private": false}
	case "POST /api/v1/accounts/999/conversations/9001/toggle_status":
		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		result = map[string]any{"id": 9001, "status": body.Status}
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Print(err)
	}
}
