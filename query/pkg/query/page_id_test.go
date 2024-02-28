package query

import (
	"encoding/json"
	"testing"
)

func TestPageId_MarshalJSON(t *testing.T) {
	var b []byte
	var err error

	p := &PageId{
		LastBlock:         19317590,
		DirectionNextPage: true,
		BlockNumber:       19317517,
		TransactionIndex:  7,
	}

	b, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	if s := string(b); s != `"QVFCV3d5WUJEY01tQVFjQUFBQT0="` {
		t.Fatal("wrong value:", s)
	}
}

func TestPageId_UnmarshalJSON(t *testing.T) {
	var s struct {
		PageId *PageId `json:"pageId"`
	}
	str := `{"pageId": "QVFCV3d5WUJEY01tQVFjQUFBQT0="}`

	if err := json.Unmarshal([]byte(str), &s); err != nil {
		t.Fatal(err)
	}

	if v := s.PageId.LastBlock; v != 19317590 {
		t.Fatal("wrong last block:", v)
	}
	if v := s.PageId.DirectionNextPage; !v {
		t.Fatal("expected DirectionNextPage to be true")
	}
	if v := s.PageId.BlockNumber; v != 19317517 {
		t.Fatal("wrong latest block:", v)
	}
	if v := s.PageId.TransactionIndex; v != 7 {
		t.Fatal("wrong latest block:", v)
	}
}
