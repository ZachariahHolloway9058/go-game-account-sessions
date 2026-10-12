package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type player struct {
	Email, Name, Password string
	Assets                []string
}
type service struct {
	mu       sync.Mutex
	players  map[string]player
	sessions map[string]string
	captcha  *infraiClient
}

func (s *service) signup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email          string
		Password       string
		Name           string
		CaptchaToken   string `json:"captcha_token"`
		WidgetRecordID string `json:"widget_record_id"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Email == "" || in.Password == "" {
		http.Error(w, "invalid signup", 400)
		return
	}
	if in.CaptchaToken != "" {
		if in.WidgetRecordID == "" {
			http.Error(w, "missing widget_record_id", 400)
			return
		}
		if err := s.captcha.verify(r.Context(), in.WidgetRecordID, in.CaptchaToken); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.players[in.Email]; ok {
		http.Error(w, "email already registered", 409)
		return
	}
	s.players[in.Email] = player{Email: in.Email, Name: in.Name, Password: in.Password}
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (s *service) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "invalid login", 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.players[in.Email]
	if !ok || p.Password != in.Password {
		http.Error(w, "invalid credentials", 401)
		return
	}
	s.sessions[in.Email] = "session-" + time.Now().UTC().Format("20060102150405.000000000")
	json.NewEncoder(w).Encode(map[string]string{"session_id": s.sessions[in.Email]})
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	c := &infraiClient{key: key, http: http.DefaultClient}
	s := &service{players: map[string]player{}, sessions: map[string]string{}, captcha: c}
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", s.signup)
	mux.HandleFunc("/login", s.login)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
