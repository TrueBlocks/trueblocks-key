package query

import (
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

func TestRpcRequest_LambdaPayload(t *testing.T) {
	r := &RpcRequest{
		Id:     1,
		Method: "test_method",
	}
	SetParams(r, []RpcGetAppearancesParam{
		{
			Address: "0x0000000000000281526004018083600019166000",
			PerPage: 16,
		},
	})

	result, err := r.LambdaPayload()
	if err != nil {
		t.Fatal(err)
	}

	if result != `{"body": "{\"id\":1,\"method\":\"test_method\",\"params\":[{\"address\":\"0x0000000000000281526004018083600019166000\",\"perPage\":16}]}"}` {
		t.Fatal("wrong value:", result)
	}
}

func TestRpcRequest_SetPageId(t *testing.T) {
	p := &PageId{
		LastSeen: database.Appearance{1000, 1},
	}

	param := RpcGetAppearancesParam{
		Address: "0x0000000000000281526004018083600019166000",
	}

	if err := param.SetPageId(PageIdNoSpecial, p); err != nil {
		t.Fatal(err)
	}
	_, pageId, err := param.PageIdValue()
	if err != nil {
		t.Fatal(err)
	}
	if pageId == nil {
		t.Fatal("pageId is nil")
	}
	if v := pageId.LastSeen; v != p.LastSeen {
		t.Fatal("wrong value:", v)
	}

	// Special

	if err := param.SetPageId(PageIdEarliest, nil); err != nil {
		t.Fatal(err)
	}
	special, pageId, err := param.PageIdValue()
	if err != nil {
		t.Fatal(err)
	}
	if pageId != nil {
		t.Fatal("expected pageId to be nil")
	}
	if special != PageIdEarliest {
		t.Fatal("wrong value:", special)
	}
}
