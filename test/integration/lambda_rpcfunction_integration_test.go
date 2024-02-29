//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
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
		BlockNumber:      1,
		TransactionIndex: 5,
	}
	if err = appearance.Insert(context.TODO(), dbConn, address); err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse[[]database.Appearance]{}

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

	if l := len(response.Result.Data); l != 1 {
		t.Fatal("wrong result count:", l)
	}
	if bn := response.Result.Data[0].BlockNumber; bn != appearance.BlockNumber {
		t.Fatal("wrong block number:", bn)
	}
	if txid := response.Result.Data[0].TransactionIndex; txid != appearance.TransactionIndex {
		t.Fatal("wrong txid:", txid)
	}

	// meta
	if response.Meta == nil {
		t.Fatal("meta is nil")
	}
	if l := response.Meta.LastIndexedBlock; l != 1 {
		t.Fatal("wrong meta LastIndexedBlock")
	}
	if a := response.Meta.Address; a != address {
		t.Fatal("wrong meta address")
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

	if l := len(response.Result.Data); l != 0 {
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

	// request = &query.RpcRequest{
	// 	Id:     1,
	// 	Method: "tb_getAppearances",
	// 	Params: []query.RpcRequestParams{
	// 		{
	// 			Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
	// 			PageId:  -1,
	// 		},
	// 	},
	// }
	// output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	// t.Logf("result: %+v", response)
	// helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: invalid PerPage

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
				PerPage: -1,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect perPage")

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
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect perPage")

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

	countResponse := &query.RpcResponse[*int]{}

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

	t.Log(string(output.Payload))
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, countResponse)

	if c := countResponse.Result.Data; *c != 1 {
		t.Fatal("wrong count:", c)
	}

	// meta
	if countResponse.Meta == nil {
		t.Fatal("meta is nil")
	}
	if l := countResponse.Meta.LastIndexedBlock; l != 1 {
		t.Fatal("wrong meta LastIndexedBlock")
	}
	if a := countResponse.Meta.Address; a != address {
		t.Fatal("wrong meta address")
	}

	// Last indexed block

	statusResponse := &query.RpcResponse[*database.Status]{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_status",
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, statusResponse)

	if l := statusResponse.Result.Meta.LastIndexedBlock; l != 1 {
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
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 20},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 19},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 18},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 17},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 16},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 15},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 14},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 13},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 12},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 11},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 10},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 9},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 8},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 7},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 6},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 5},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 4},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 3},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 2},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 3001234, TransactionIndex: 1},
	}
	if err = database.InsertAppearanceBatch(context.TODO(), dbConn, appearances); err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse[[]database.Appearance]{}
	perPage := 10
	maxIters := len(appearances) / perPage

	// Check basic pagination

	var nextPageId *query.PageId
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: 10,
				},
			},
		}
		if nextPageId != nil {
			if err := request.SetPageId(query.PageIdNoSpecial, nextPageId); err != nil {
				t.Fatal(err)
			}
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); l != request.Parameters().PerPage {
			t.Fatal(i, "-- wrong result count:", l)
		}

		pa := make([]database.Appearance, 0, len(response.Result.Data))
		startIndex := i * perPage
		endIndex := startIndex + perPage
		for _, item := range appearances[startIndex:endIndex] {
			pa = append(pa, database.Appearance{
				BlockNumber:      item.BlockNumber,
				TransactionIndex: item.TransactionIndex,
			})
		}
		if r := response.Result.Data; !reflect.DeepEqual(r, pa) {
			t.Fatal(i, "-- wrong result:", r)
		}

		if i == 0 && response.Result.Meta.PreviousPageId != nil {
			t.Fatal("first page returned PreviousPageId")
		}

		nextPageId = response.Meta.NextPageId
	}

	// Check items

	nextPageId = nil
	var pagingResults = make([]database.Appearance, 0, len(appearances))
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: perPage,
				},
			},
		}
		if err := request.SetPageId(query.PageIdNoSpecial, nextPageId); err != nil {
			t.Fatal(err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); l != perPage {
			t.Fatal(i, "-- wrong page len:", l)
		}
		pagingResults = append(pagingResults, response.Result.Data...)

		nextPageId = response.Meta.NextPageId
	}

	if l := len(pagingResults); l != len(appearances) {
		t.Fatal("wrong result length", l, "expected", len(appearances))
	}

	for index, pa := range pagingResults {
		if bn := pa.BlockNumber; bn != appearances[index].BlockNumber {
			t.Fatal("wrong block number", bn, "expected", appearances[index].BlockNumber)
		}
		if txid := pa.TransactionIndex; txid != appearances[index].TransactionIndex {
			t.Fatal("wrong txid", txid, "expected", appearances[index].TransactionIndex)
		}
	}

	// Check pageIds

	nextPageId = nil
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: perPage,
				},
			},
		}
		if err := request.SetPageId(query.PageIdNoSpecial, nextPageId); err != nil {
			t.Fatal(err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)
		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if i == 0 {
			if response.Result.Meta.PreviousPageId != nil {
				t.Fatal("expected no PreviousPageId on the first page")
			}
		}
		if i == maxIters-1 {
			if response.Result.Meta.NextPageId != nil {
				t.Fatal("expected no NextPageId on the last page")
			}
		}
		nextPageId = response.Result.Meta.NextPageId
	}

	// Check pageId = "" same as pageId = "latest"

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: appearances[0].Address,
				PerPage: perPage,
			},
		},
	}
	if err := request.SetPageId(query.PageIdLatest, nil); err != nil {
		t.Fatal(err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	latestApps := response.Result.Data

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address: appearances[0].Address,
				PerPage: perPage,
				PageId:  []byte(""),
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	if !reflect.DeepEqual(latestApps, response.Result.Data) {
		t.Fatal("wrong results")
	}

	// Check going backwards with "earliest"

	var prevPageId *query.PageId
	pagingResults = make([]database.Appearance, 0, len(appearances))
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: []query.RpcRequestParams{
				{
					Address: appearances[0].Address,
					PerPage: perPage,
				},
			},
		}
		var pageIdSpecial query.PageIdSpecial
		var pageId *query.PageId
		if i == 0 {
			pageIdSpecial = query.PageIdEarliest
		} else {
			pageId = prevPageId
		}
		if err := request.SetPageId(pageIdSpecial, pageId); err != nil {
			t.Fatal(err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); l != perPage {
			t.Fatal(i, "-- wrong page len:", l)
		}
		pagingResults = append(pagingResults, response.Result.Data...)

		prevPageId = response.Meta.PreviousPageId
	}

	if l := len(pagingResults); l != len(appearances) {
		t.Fatal("wrong result length", l, "expected", len(appearances))
	}

	expected := []database.Appearance{
		{BlockNumber: 3001234, TransactionIndex: 10},
		{BlockNumber: 3001234, TransactionIndex: 9},
		{BlockNumber: 3001234, TransactionIndex: 8},
		{BlockNumber: 3001234, TransactionIndex: 7},
		{BlockNumber: 3001234, TransactionIndex: 6},
		{BlockNumber: 3001234, TransactionIndex: 5},
		{BlockNumber: 3001234, TransactionIndex: 4},
		{BlockNumber: 3001234, TransactionIndex: 3},
		{BlockNumber: 3001234, TransactionIndex: 2},
		{BlockNumber: 3001234, TransactionIndex: 1},
		{BlockNumber: 4053179, TransactionIndex: 20},
		{BlockNumber: 4053179, TransactionIndex: 19},
		{BlockNumber: 4053179, TransactionIndex: 18},
		{BlockNumber: 4053179, TransactionIndex: 17},
		{BlockNumber: 4053179, TransactionIndex: 16},
		{BlockNumber: 4053179, TransactionIndex: 15},
		{BlockNumber: 4053179, TransactionIndex: 14},
		{BlockNumber: 4053179, TransactionIndex: 13},
		{BlockNumber: 4053179, TransactionIndex: 12},
		{BlockNumber: 4053179, TransactionIndex: 11},
	}

	if !reflect.DeepEqual(expected, pagingResults) {
		t.Fatal("wrong results")
	}

	// lastBlock custom

	customLastBlock := 3001234
	b, err := json.Marshal(customLastBlock)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(b)
	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: []query.RpcRequestParams{
			{
				Address:   appearances[0].Address,
				PerPage:   perPage,
				LastBlock: &raw,
			},
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	if l := len(response.Result.Data); l != perPage {
		t.Fatal("wrong page len:", l)
	}
	if response.Result.Data[0].BlockNumber != uint32(customLastBlock) {
		t.Fatal("wrong block number")
	}
}
