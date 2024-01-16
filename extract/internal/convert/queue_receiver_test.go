package convert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

func TestQueueReceiver_SendBatch(t *testing.T) {
	batch := []queueItem.Appearance{
		{
			Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			BlockNumber:     18540199,
			TransactionId:   1,
			BlockRangeStart: 18540000,
			BlockRangeEnd:   18540200,
		},
		{
			Address:         "0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5",
			BlockNumber:     18540198,
			TransactionId:   12,
			BlockRangeStart: 18540000,
			BlockRangeEnd:   18540200,
		},
		{
			Address:         "0x0f64881A3BFA0789dbE57F2721098395A79B1ec6",
			BlockNumber:     18540197,
			TransactionId:   13,
			BlockRangeStart: 18540000,
			BlockRangeEnd:   18540200,
		},
	}
	results := make([]*database.Appearance, 0)
	var mutex sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.Method != "POST" {
			t.Fatal("wrong method", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		app := &database.Appearance{}
		if err := json.Unmarshal(body, app); err != nil {
			t.Fatal(err)
		}
		mutex.Lock()
		defer mutex.Unlock()
		results = append(results, app)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	qr := &QueueReceiver{
		InsertUrl: ts.URL,
	}

	if err := qr.SendBatch(batch); err != nil {
		t.Fatal(err)
	}

	expected := make([]*database.Appearance, 0, len(batch))
	for _, item := range batch {
		expected = append(expected, &database.Appearance{
			BlockNumber:   item.BlockNumber,
			TransactionId: item.TransactionId,
		})
	}

	if !reflect.DeepEqual(expected, results) {
		t.Fatal("wrong results")
	}
}
