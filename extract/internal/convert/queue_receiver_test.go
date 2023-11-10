package convert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	database "trueblocks.io/database/pkg"
)

func TestQueueReceiver_SendBatch(t *testing.T) {
	batch := []*database.Appearance{
		{
			Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			BlockNumber:     18540199,
			TransactionId:   1,
			BlockRangeStart: 18540000,
			BlockRangeEnd:   18540200,
		},
		{
			Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			BlockNumber:     18540199,
			TransactionId:   2,
			BlockRangeStart: 18540000,
			BlockRangeEnd:   18540200,
		},
		{
			Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			BlockNumber:     18540199,
			TransactionId:   3,
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

	if !reflect.DeepEqual(batch, results) {
		t.Fatal("wrong results")
	}
}
