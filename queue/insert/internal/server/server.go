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

var qu *queue.Queue

func Start(port int, q *queue.Queue) (err error) {
	qu = q
	http.HandleFunc("/add", addHandler)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return
}

func addHandler(w http.ResponseWriter, r *http.Request) {
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

	msgId, err := qu.Add(app)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(200)
	w.Write([]byte(msgId))
}
