package queue

import (
	"reflect"
	"testing"

	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/queue/queuetest"
)

func TestQueueAdd(t *testing.T) {
	mock := &queuetest.MockQueue{}
	q, err := NewQueue(mock)
	if err != nil {
		t.Fatal(err)
	}

	app := &queueItem.Appearance{
		Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:     17742858,
		TransactionId:   15,
		BlockRangeStart: 17740000,
		BlockRangeEnd:   17742858,
	}
	if _, err := q.AddAppearance(app); err != nil {
		t.Fatal(err)
	}

	if l := mock.Len(); l != 1 {
		t.Fatal("wrong queue length:", l)
	}

	item := mock.GetAppearances(0)
	if !reflect.DeepEqual(item, app) {
		t.Fatalf("expected %+v but got %+v", app, item)
	}

	chunk := &queueItem.Chunk{
		Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
		Range:  "1000-2000",
		Author: "test",
	}

	if _, err := q.AddChunk(chunk); err != nil {
		t.Fatal(err)
	}

	if l := mock.Len(); l != 2 {
		t.Fatal("wrong queue length:", l)
	}

	chunkItem := mock.GetChunks(0)
	if !reflect.DeepEqual(chunkItem, chunk) {
		t.Fatalf("expected %+v but got %+v", chunk, chunkItem)
	}
}

func TestQueueAddBatch(t *testing.T) {
	mock := &queuetest.MockQueue{}
	q, err := NewQueue(mock)
	if err != nil {
		t.Fatal(err)
	}

	apps := []*queueItem.Appearance{
		{
			Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			BlockNumber:     17742858,
			TransactionId:   15,
			BlockRangeStart: 17740000,
			BlockRangeEnd:   17742858,
		},
	}
	if err := q.AddAppearanceBatch(apps); err != nil {
		t.Fatal(err)
	}

	if l := mock.Len(); l != 1 {
		t.Fatal("wrong queue length:", l)
	}

	item := mock.GetAppearances(0)
	if !reflect.DeepEqual(item, apps[0]) {
		t.Fatalf("expected %+v but got %+v", item, apps[0])
	}

	chunks := []*queueItem.Chunk{
		{
			Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
			Range:  "1000-2000",
			Author: "test",
		},
	}

	if err := q.AddChunkBatch(chunks); err != nil {
		t.Fatal(err)
	}

	if l := mock.Len(); l != 2 {
		t.Fatal("wrong queue length:", l)
	}

	chunkItem := mock.GetChunks(0)
	if !reflect.DeepEqual(chunkItem, chunks[0]) {
		t.Fatalf("expected %+v but got %+v", chunks[0], chunkItem)
	}
}
