//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/TrueBlocks/trueblocks-key/test/dbtest"
	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// rawPayload is used to send a request with totally wrong parameters
type rawPayload string

func (r rawPayload) LambdaPayload() (string, error) {
	return string(r), nil
}

func TestLambdaRpcFunctionRequests(t *testing.T) {
	dbConn, done, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	// Prepate test data
	address := "0x0000000000000281526004018083600019166000"
	appearance := &database.Appearance{
		BlockNumber:   1,
		TransactionId: 5,
	}
	if err = appearance.Insert(context.TODO(), dbConn, address); err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcAppearancesResponse{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: address,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	t.Logf("result: %+v", response)

	if l := len(response.Result); l != 1 {
		t.Fatal("wrong result count:", l)
	}
	if bn := response.Result[0].BlockNumber; bn != appearance.BlockNumber {
		t.Fatal("wrong block number:", bn)
	}
	if txid := response.Result[0].TransactionId; txid != appearance.TransactionId {
		t.Fatal("wrong txid:", txid)
	}

	// Valid request, no appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	t.Logf("result: %+v", response)

	if l := len(response.Result); l != 0 {
		t.Fatal("wrong result count:", l)
	}

	// Invalid request: exactly 1 parameter object required

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "exactly 1 parameter object required")

	// Invalid request: invalid address

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: "0000000000000281526004018083600019166000",
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect address")

	// Invalid request: invalid page

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				Page:    -1,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: invalid PerPage

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				Page:    10,
				PerPage: -1,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: params out of range

	outOfIntRange := big.NewInt(0)
	// Set outOfIntRange value to max value of int as mentioned here: https://stackoverflow.com/questions/6878590/the-maximum-value-for-an-int-type-in-go
	outOfIntRange.SetString(fmt.Sprint(int((^uint(0))>>1)), 10)
	// Now make it out of range
	outOfIntRange.Add(outOfIntRange, big.NewInt(1))
	rp := rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"tb_getAppearances\",\"params\":{\"address\":\"0x0000000000000281526004018083600019166000\",\"page\":8,\"perPage\":%s}}"}`, outOfIntRange.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "invalid JSON")

	// Invalid request: insane number as parameter

	insane := big.NewInt(1 << 60)
	rp = rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"tb_getAppearances\",\"params\":[{\"address\":\"0x0000000000000281526004018083600019166000\",\"page\":8,\"perPage\":%s}]}"}`, insane.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: invalid method

	request = &query.RpcRequest{
		Id:     1,
		Method: "invalid",
		Params: []query.RpcRequestParams{
			{
				Address: address,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "invalid method")

	// Count

	countResponse := &query.RpcCountResponse{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearanceCount",
		Params: []query.RpcRequestParams{
			{
				Address: address,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, countResponse)

	if c := countResponse.Result; c != 1 {
		t.Fatal("wrong count:", c)
	}

	// Last indexed block

	lastIndexedBlockResponse := &query.RpcLastIndexedBlockResponse{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_lastIndexedBlock",
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, lastIndexedBlockResponse)

	if l := lastIndexedBlockResponse.Result; l != 1 {
		t.Fatal("wrong max indexed block:", l)
	}
}

func TestLambdaRpcFunctionPagination(t *testing.T) {
	dbConn, done, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	// Prepate test data
	appearances := []queueItem.Appearance{
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 1},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 2},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 3},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 4},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 5},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 6},
	}
	if err = database.InsertAppearanceBatch(context.TODO(), dbConn, appearances); err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcAppearancesResponse{}

	// Check basic pagination

	for i := 0; i < len(appearances); i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: 1,
					Page:    i,
				},
			},
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result); l != 1 {
			t.Fatal(i, "-- wrong result count:", l)
		}

		pa := database.Appearance{
			BlockNumber:   appearances[i].BlockNumber,
			TransactionId: appearances[i].TransactionId,
		}
		if r := response.Result; !reflect.DeepEqual(r, []database.Appearance{pa}) {
			t.Fatal(i, "-- wrong result:", r)
		}
	}

	// Check items

	var pagingResults = make([]database.Appearance, 0, len(appearances))
	perPage := 3
	for i := 0; i < (len(appearances) / perPage); i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: perPage,
					Page:    i,
				},
			},
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result); l != 3 {
			t.Fatal(i, "-- wrong page len:", l)
		}
		pagingResults = append(pagingResults, response.Result...)
	}

	if l := len(pagingResults); l != len(appearances) {
		t.Fatal("wrong result length", l, "expected", len(appearances))
	}

	for index, pa := range pagingResults {
		// if addr := pa.Address; addr != appearances[index].Address {
		// 	t.Fatal("wrong address", addr, "expected", appearances[index].Address)
		// }
		if bn := pa.BlockNumber; bn != appearances[index].BlockNumber {
			t.Fatal("wrong block number", bn, "expected", appearances[index].BlockNumber)
		}
		if txid := pa.TransactionId; txid != appearances[index].TransactionId {
			t.Fatal("wrong txid", txid, "expected", appearances[index].TransactionId)
		}
	}
}
