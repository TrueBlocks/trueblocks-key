package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	coreNotify "github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/notify"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/queue"
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
	http.HandleFunc("/batch", s.batchHandler)

	url := fmt.Sprintf(":%d", port)
	fmt.Println("Listening:", url)
	err = http.ListenAndServe(url, nil)
	return
}

func (s *Server) Error(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func (s *Server) addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
		return
	}

	// app := &appearance.Appearance{}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.Error(w, 500, err)
		return
	}

	notificationType, err := readNotificationType(b)
	if err != nil {
		s.Error(w, 400, err)
	}

	var msgId string
	switch Message(notificationType) {
	case MessageAppearance:
		s.Error(w, 400, fmt.Errorf("appearance type is only supported in batches"))
		return
	case MessageChunkWritten:
		notification := &coreNotify.Notification[coreNotify.NotificationPayloadChunkWritten]{}
		if err := json.Unmarshal(b, notification); err != nil {
			fmt.Println(err)
			s.Error(w, 500, err)
			return
		}
		chunk := &queueItem.Chunk{
			Cid:    notification.Payload.Cid,
			Range:  notification.Payload.Range,
			Author: notification.Payload.Author,
		}
		msgId, err = s.qu.AddChunk(chunk)
		if err != nil {
			s.Error(w, 400, err)
			return
		}
		log.Println("Added notification of type ChunkWritten")
	case MessageStageUpdated:
		w.WriteHeader(208)
		return
	default:
		s.Error(w, 400, fmt.Errorf("unknown notification type: %s", notificationType))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(msgId))
}

func (s *Server) batchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(400)
		return
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.Error(w, 500, err)
		return
	}

	notificationType, err := readNotificationType(b)
	if err != nil {
		s.Error(w, 400, err)
	}

	switch Message(notificationType) {
	case MessageAppearance:
		notification := &coreNotify.Notification[[]coreNotify.NotificationPayloadAppearance]{}
		if err := json.Unmarshal(b, notification); err != nil {
			s.Error(w, 500, err)
			return
		}
		apps, err := Appearances(notification)
		if err != nil {
			s.Error(w, 500, err)
			return
		}
		if err := s.qu.AddAppearanceBatch(apps); err != nil {
			s.Error(w, 400, err)
			return
		}
		log.Printf("Batch added %d notifications of type Appearance\n", len(apps))
	case MessageChunkWritten:
		notification := &coreNotify.Notification[[]coreNotify.NotificationPayloadChunkWritten]{}
		if err := json.Unmarshal(b, notification); err != nil {
			s.Error(w, 500, err)
			return
		}
		chunks := make([]*queueItem.Chunk, 0, len(notification.Payload))
		for _, item := range notification.Payload {
			chunks = append(chunks, &queueItem.Chunk{
				Cid:    item.Cid,
				Range:  item.Range,
				Author: item.Author,
			})
		}
		if err := s.qu.AddChunkBatch(chunks); err != nil {
			s.Error(w, 400, err)
			return
		}
		log.Printf("Batch added %d notifications of type ChunkWritten\n", len(chunks))
	case MessageStageUpdated:
		w.WriteHeader(208)
		return
	default:
		s.Error(w, 400, fmt.Errorf("unknown notification type: %s", notificationType))
		return
	}

	w.WriteHeader(200)
}

func readNotificationType(b []byte) (string, error) {
	var part struct {
		Msg string `json:"msg"`
	}

	if err := json.Unmarshal(b, &part); err != nil {
		return "", err
	}

	return part.Msg, nil
}
