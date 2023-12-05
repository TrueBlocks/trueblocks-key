package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/queue"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/queue/queuetest"
)

var testPort = 9999

func TestServer_Add(t *testing.T) {
	mockQueue := &queuetest.MockQueue{}
	q, _ := queue.NewQueue(mockQueue)
	svr := New(q)
	ts := httptest.NewServer(http.HandlerFunc(svr.addHandler))
	defer ts.Close()

	expected := &appearance.Appearance{
		Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:     18540199,
		TransactionId:   1,
		BlockRangeStart: 18540000,
		BlockRangeEnd:   18540200,
	}
	encoded, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(ts.URL, "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("wrong status code:", res.StatusCode)
	}

	if result := mockQueue.Get(0); !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %+v, but got %+v", expected, result)
	}
}

func TestServer_AddBatch(t *testing.T) {
	mockQueue := &queuetest.MockQueue{}
	q, _ := queue.NewQueue(mockQueue)
	svr := New(q)
	ts := httptest.NewServer(http.HandlerFunc(svr.batchHandler))
	defer ts.Close()

	payload := Notification{
		Payload: NotificationPayload{
			{
				Address:          "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				BlockNumber:      "18540199",
				TransactionIndex: 1,
			},
			{
				Address:          "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				BlockNumber:      "18540199",
				TransactionIndex: 2,
			},
			{
				Address:          "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				BlockNumber:      "18540199",
				TransactionIndex: 3,
			},
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(ts.URL, "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("wrong status code:", res.StatusCode)
	}

	apps, err := payload.Appearances()
	if err != nil {
		t.Fatal(err)
	}
	expected := make([]*appearance.Appearance, 0, len(apps))
	for _, appearance := range apps {
		expected = append(expected, appearance)
	}

	for index, exp := range expected {
		if result := mockQueue.Get(index); !reflect.DeepEqual(result, exp) {
			t.Fatalf("expected %+v, but got %+v", exp, result)
		}
	}
}
