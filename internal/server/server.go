package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/VOVAN1993/poker_hand/internal/hander"
)

type Server struct {
	handManager hander.HandManager
}

func NewServer(handManager hander.HandManager) *Server {
	return &Server{handManager: handManager}
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) error {
	response, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, _ = w.Write(response)

	return nil
}

func (s *Server) tournamentHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		t, err := s.handManager.GetTournament(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Server error: %s", err)))
			return
		}
		RespondJSON(w, http.StatusOK, t)
	}
}

func (s *Server) tournamentsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := s.handManager.ListTournaments(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Server error: %s", err)))
			return
		}
		RespondJSON(w, http.StatusOK, ts)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func (s *Server) Start() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/tournamets", s.tournamentsHandler())
	http.HandleFunc("/tournamets/{id}", s.tournamentHandler())
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
