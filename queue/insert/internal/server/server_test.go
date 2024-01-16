package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
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

	expected := &queueItem.Chunk{
		Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
		Range:  "1000-2000",
		Author: "test",
	}
	n := &Notification[NotificationPayloadChunkWritten]{
		Msg: MessageChunkWritten,
		Payload: NotificationPayloadChunkWritten{
			Cid:    expected.Cid,
			Range:  expected.Range,
			Author: expected.Author,
		},
	}
	encoded, err := json.Marshal(n)
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

	if result := mockQueue.GetChunks(0); !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %+v, but got %+v", expected, result)
	}
}

func TestServer_AddBatch(t *testing.T) {
	mockQueue := &queuetest.MockQueue{}
	q, _ := queue.NewQueue(mockQueue)
	svr := New(q)
	ts := httptest.NewServer(http.HandlerFunc(svr.batchHandler))
	defer ts.Close()

	payload := Notification[[]NotificationPayloadAppearance]{
		Msg: MessageAppearance,
		Payload: []NotificationPayloadAppearance{
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
	expected := make([]*queueItem.Appearance, 0, len(apps))
	for _, appearance := range apps {
		expected = append(expected, appearance)
	}

	for index, exp := range expected {
		if result := mockQueue.GetAppearances(index); !reflect.DeepEqual(result, exp) {
			t.Fatalf("expected %+v, but got %+v", exp, result)
		}
	}
}

func TestServer_AddBatch_Chunks(t *testing.T) {
	mockQueue := &queuetest.MockQueue{}
	q, _ := queue.NewQueue(mockQueue)
	svr := New(q)
	ts := httptest.NewServer(http.HandlerFunc(svr.batchHandler))
	defer ts.Close()

	payload := Notification[[]NotificationPayloadChunkWritten]{
		Msg: MessageChunkWritten,
		Payload: []NotificationPayloadChunkWritten{
			{
				Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
				Range:  "1000-2000",
				Author: "test1",
			},
			{
				Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
				Range:  "1000-2000",
				Author: "test2",
			},
			{
				Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
				Range:  "1000-2000",
				Author: "test3",
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

	expected := make([]*queueItem.Chunk, 0, len(payload.Payload))
	for _, chunk := range payload.Payload {
		expected = append(expected, &queueItem.Chunk{
			Cid:    chunk.Cid,
			Range:  chunk.Range,
			Author: chunk.Author,
		})
	}

	for index, exp := range expected {
		if result := mockQueue.GetChunks(index); !reflect.DeepEqual(result, exp) {
			t.Fatalf("expected %+v, but got %+v", exp, result)
		}
	}
}
