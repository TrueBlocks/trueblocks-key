package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"trueblocks.io/queue/consume/pkg/appearance"
	"trueblocks.io/queue/insert/internal/queue"
)

// This will be removed when Scraper supports Notify

type Server struct {
	qu *queue.Queue
}

func New(qu *queue.Queue) *Server {
	return &Server{
		qu,
	}
}

func (s *Server) Start(port int) (err error) {
	http.HandleFunc("/add", s.addHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return
}

func (s *Server) addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
	}

	app := &appearance.Appearance{}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
	if err := json.Unmarshal(b, app); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

	msgId, err := s.qu.Add(app)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(200)
	w.Write([]byte(msgId))
}
