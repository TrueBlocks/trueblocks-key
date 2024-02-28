package query

import (
	"encoding/json"
	"reflect"
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

func TestPageId_MarshalJSON(t *testing.T) {
	var b []byte
	var err error

	p := &PageId{
		LastBlock:         19317590,
		DirectionNextPage: true,
		LastSeen:          database.Appearance{BlockNumber: 19317517, TransactionIndex: 7},
		LatestInSet:       database.Appearance{BlockNumber: 19317590, TransactionIndex: 10},
		EarliestInSet:     database.Appearance{BlockNumber: 100000, TransactionIndex: 126},
	}

	b, err = json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	if s := string(b); s != `"QVZiREpnRU53eVlCQndBQUFGYkRKZ0VLQUFBQW9JWUJBSDRBQUFBPQ=="` {
		t.Fatal("wrong value:", s)
	}
}

func TestPageId_UnmarshalJSON(t *testing.T) {
	var s struct {
		PageId *PageId `json:"pageId"`
	}
	str := `{"pageId": "QVZiREpnRU53eVlCQndBQUFGYkRKZ0VLQUFBQW9JWUJBSDRBQUFBPQ=="}`

	if err := json.Unmarshal([]byte(str), &s); err != nil {
		t.Fatal(err)
	}

	if v := s.PageId.LastBlock; v != 19317590 {
		t.Fatal("wrong last block:", v)
	}
	if v := s.PageId.DirectionNextPage; !v {
		t.Fatal("expected DirectionNextPage to be true")
	}
	if v := s.PageId.LastSeen; !reflect.DeepEqual(v, database.Appearance{BlockNumber: 19317517, TransactionIndex: 7}) {
		t.Fatal("wrong LastSeen:", v)
	}
	if v := s.PageId.LatestInSet; !reflect.DeepEqual(v, database.Appearance{BlockNumber: 19317590, TransactionIndex: 10}) {
		t.Fatal("wrong LatestInSet:", v)
	}
	if v := s.PageId.EarliestInSet; !reflect.DeepEqual(v, database.Appearance{BlockNumber: 100000, TransactionIndex: 126}) {
		t.Fatal("wrong EarliestInSet:", v)
	}
}
