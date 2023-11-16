package convert

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
)

// QueueReceiver uses HTTP to send appearances to `insert` tool running
type QueueReceiver struct {
	// InsertUrl is where `insert` tool is running
	InsertUrl string
	// MaxConnections is number of parallel requests to `insert` tool
	MaxConnections int

	// sendMutex causes the queue to block while it's sending requests,
	// so any parallel calls to SendBatch() will wait
	sendMutex sync.Mutex
}

// SendBatch sends given batch of items to `insert` tool
func (q *QueueReceiver) SendBatch(batch []*database.Appearance) (err error) {
	q.sendMutex.Lock()
	defer q.sendMutex.Unlock()

	encodedItems := make(chan []byte, q.MaxConnections)
	errs := make(chan error)

	go q.encodeBatch(batch, encodedItems, errs)
	go func() {
		q.send(encodedItems, errs)
		close(errs)
	}()

	if err = <-errs; err != nil {
		return err
	}

	return
}

func (q *QueueReceiver) encodeBatch(batch []*database.Appearance, results chan<- []byte, errs chan<- error) {
	defer close(results)
	for _, item := range batch {
		item := item
		app := &appearance.Appearance{
			Address:         item.Address,
			TransactionId:   item.TransactionId,
			BlockNumber:     item.BlockNumber,
			BlockRangeStart: item.BlockRangeStart,
			BlockRangeEnd:   item.BlockRangeEnd,
		}
		var encoded []byte
		encoded, err := json.Marshal(app)
		if err != nil {
			errs <- err
			return
		}
		results <- encoded
	}
}

func (q *QueueReceiver) send(encoded <-chan []byte, errs chan<- error) {
	for item := range encoded {
		if err := q.sendOne(item); err != nil {
			errs <- err
			return
		}
	}
}

func (q *QueueReceiver) sendOne(encoded []byte) (err error) {
	var resp *http.Response
	resp, err = http.Post(q.InsertUrl+"/add", "application/json", bytes.NewReader(encoded))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return errors.New(string(bodyBytes))
	}
	return
}
