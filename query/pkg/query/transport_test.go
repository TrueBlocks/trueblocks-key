package query

import "testing"

func TestRpcRequest_LambdaPayload(t *testing.T) {
	r := &RpcRequest{
		Id:     1,
		Method: "test_method",
		Params: []RpcRequestParams{
			{
				Address: "0x0000000000000281526004018083600019166000",
				PageId:  nil,
				PerPage: 16,
			},
		},
	}

	result, err := r.LambdaPayload()
	if err != nil {
		t.Fatal(err)
	}

	if result != `{"body": "{\"id\":1,\"method\":\"test_method\",\"params\":[{\"address\":\"0x0000000000000281526004018083600019166000\",\"perPage\":16}]}"}` {
		t.Fatal("wrong value:", result)
	}
}
