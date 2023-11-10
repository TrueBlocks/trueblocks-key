package queue

import (
	"testing"

	"trueblocks.io/queue/consume/pkg/appearance"
	"trueblocks.io/queue/insert/internal/queue/queuetest"
)

func TestQueueAdd(t *testing.T) {
	mock := &queuetest.MockQueue{}
	q, err := NewQueue(mock)
	if err != nil {
		t.Fatal(err)
	}

	app := &appearance.Appearance{
		Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:     17742858,
		TransactionId:   15,
		BlockRangeStart: 17740000,
		BlockRangeEnd:   17742858,
	}
	q.Add(app)

	if l := mock.Len(); l != 1 {
		t.Fatal("wrong queue length:", l)
	}

	item := mock.Get(0)
	if item.AppearanceId != app.AppearanceId {
		t.Fatal("wrong AppearanceId")
	}
	if item.BlockNumber != app.BlockNumber {
		t.Fatal("wrong BlockNumber")
	}
	if item.TransactionId != app.TransactionId {
		t.Fatal("wrong TransactionId")
	}
	if item.Address != app.Address {
		t.Fatal("wrong Address")
	}
	if item.BlockRangeStart != app.BlockRangeStart {
		t.Fatal("wrong BlockRangeStart")
	}
	if item.BlockRangeEnd != app.BlockRangeEnd {
		t.Fatal("wrong BlockRangeEnd")
	}

}
