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

func TestPageId_Errors(t *testing.T) {
	var s struct {
		PageId *PageId `json:"pageId"`
	}
	str := `{"pageId": "__invalid__"}`

	err := json.Unmarshal([]byte(str), &s)
	if err == nil {
		t.Fatal("expected error")
	}
	t.Fatal(err)
}

func TestPageIdSpecial_FromBytes(t *testing.T) {
	var p PageIdSpecial

	type args struct {
		b []byte
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		value PageIdSpecial
	}{
		{
			name: "invalid value",
			args: args{b: []byte("invalid")},
			want: false,
		},
		{
			name: "no value",
			args: args{b: []byte(PageIdNoSpecial)},
			want: false,
		},
		{
			name:  "latest",
			args:  args{b: []byte(PageIdLatest)},
			want:  true,
			value: PageIdLatest,
		},
		{
			name:  "earliest",
			args:  args{b: []byte(PageIdEarliest)},
			want:  true,
			value: PageIdEarliest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.FromBytes(tt.args.b)
			if got != tt.want {
				t.Errorf("PageIdSpecial.FromBytes() = %v, want %v", got, tt.want)
			}
			if p != tt.value {
				t.Error("wanted value", tt.value, "got", p)
			}
		})
	}
}
