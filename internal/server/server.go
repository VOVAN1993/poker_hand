package server

import (
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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func (s *Server) Start() {
	http.HandleFunc("/", helloHandler)
	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
