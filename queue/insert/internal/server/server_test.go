package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"trueblocks.io/queue/consume/pkg/appearance"
	"trueblocks.io/queue/insert/internal/queue"
	"trueblocks.io/queue/insert/internal/queue/queuetest"
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

	expected.SetAppearanceId()

	if result := mockQueue.Get(0); !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %+v, but got %+v", expected, result)
	}
}
