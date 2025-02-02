//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
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
	address := "0x74DF56727D04F6f30c9f52D6CCC1eBfb6C93F687"
	appearance := &database.Appearance{
		BlockNumber:      1,
		TransactionIndex: 5,
	}
	if err = appearance.Insert(context.TODO(), dbConn, address); err != nil {
		t.Fatal("inserting test data:", err)
	}
	appearanceBlockNumber := "1"
	appearanceTransactionIndex := "5"

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse[[]database.PublicAppearance]{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			{
				Address: address,
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	t.Logf("result: %+v", response)

	if l := len(response.Result.Data); l != 1 {
		t.Fatal("wrong result count:", l)
	}
	if bn := response.Result.Data[0].BlockNumber; bn != appearanceBlockNumber {
		t.Fatal("wrong block number:", bn)
	}
	if txid := response.Result.Data[0].TransactionIndex; txid != appearanceTransactionIndex {
		t.Fatal("wrong txid:", txid)
	}

	// meta
	if response.Meta == nil {
		t.Fatal("meta is nil")
	}
	if l := response.Meta.LastIndexedBlock; l != "1" {
		t.Fatal("wrong meta LastIndexedBlock")
	}
	// Meta returns address used in query, so it won't be lower cased here
	if a := response.Meta.Address; a != address {
		t.Fatal("wrong meta address", a)
	}

	// Valid request, no appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			{
				Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
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
	}

	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "exactly 1 parameter object required")

	// Invalid request: invalid address

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			{
				Address: "0000000000000281526004018083600019166000",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect address")

	// Invalid request: invalid PerPage

	// request = &query.RpcRequest{
	// 	Id:     1,
	// 	Method: "tb_getAppearances",
	// }
	// err = query.SetParams(
	// 	request,
	// 	[]query.RpcGetAppearancesParam{
	// 		{
	// 			Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
	// 			PerPage: query.MinSafePerPage - 1,
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	t.Fatal("setting rpc request params:", err)
	// }
	// output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	// t.Logf("result: %+v", response)
	// helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect perPage")

	// Invalid request: params out of range

	outOfIntRange := big.NewInt(0)
	// Set outOfIntRange value to max value of int as mentioned here: https://stackoverflow.com/questions/6878590/the-maximum-value-for-an-int-type-in-go
	outOfIntRange.SetString(fmt.Sprint(int((^uint(0))>>1)), 10)
	// Now make it out of range
	outOfIntRange.Add(outOfIntRange, big.NewInt(1))
	rp := rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"tb_getAppearances\",\"params\":{\"address\":\"0x74df56727d04f6f30c9f52d6ccc1ebfb6c93f687\",\"page\":8,\"perPage\":%s}}"}`, outOfIntRange.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "invalid JSON")

	// Invalid request: insane number as parameter

	insane := big.NewInt(1 << 60)
	rp = rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"tb_getAppearances\",\"params\":[{\"address\":\"0x74df56727d04f6f30c9f52d6ccc1ebfb6c93f687\",\"page\":8,\"perPage\":%s}]}"}`, insane.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "incorrect perPage")

	// Invalid request: invalid method

	request = &query.RpcRequest{
		Id:     1,
		Method: "invalid",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			{
				Address: address,
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaProxyError(t, string(output.Payload), "unsupported method: invalid")

	// Bounds

	boundsResponse := &query.RpcResponse[database.PublicAppearancesDatasetBounds]{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getBounds",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			{
				Address: address,
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Log(string(output.Payload))
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, boundsResponse)

	expectedBounds := database.PublicAppearancesDatasetBounds{
		Latest: database.PublicAppearance{
			BlockNumber:      "1",
			TransactionIndex: "5",
		},
		Earliest: database.PublicAppearance{
			BlockNumber:      "1",
			TransactionIndex: "5",
		},
	}

	if b := boundsResponse.Result.Data; !reflect.DeepEqual(b, expectedBounds) {
		t.Fatalf("wrong bounds: %+v\n", b)
	}

	// meta
	if boundsResponse.Result.Meta == nil {
		t.Fatal("meta is nil")
	}
	if l := boundsResponse.Result.Meta.LastIndexedBlock; l != "1" {
		t.Fatal("wrong meta LastIndexedBlock")
	}
	if a := boundsResponse.Result.Meta.Address; a != address {
		t.Fatal("wrong meta address")
	}

	// Status

	statusResponse := &query.RpcResponse[*database.Status]{}

	// Valid request

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_status",
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, statusResponse)

	if l := statusResponse.Result.Meta.LastIndexedBlock; l != "1" {
		t.Fatal("wrong max indexed block:", l)
	}

	// AddressesInTx

	getAddressesInTxResponse := &query.RpcResponse[[]string]{}

	// Valid request

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAddressesInTransaction",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAddressesInParam{
			{
				BlockNumber:      "1",
				TransactionIndex: "5",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, getAddressesInTxResponse)

	expectedAddressesInTransaction := []string{
		"0x74df56727d04f6f30c9f52d6ccc1ebfb6c93f687",
	}

	if d := getAddressesInTxResponse.Result.Data; !reflect.DeepEqual(d, expectedAddressesInTransaction) {
		t.Fatalf("wrong result: %+v", d)
	}

	// AddressesInBlock

	getAddressesInBlockResponse := &query.RpcResponse[[]string]{}

	// Valid request

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAddressesInBlock",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAddressesInParam{
			{
				BlockNumber: "1",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, getAddressesInBlockResponse)

	expectedAddressesInTransaction = []string{
		"0x74df56727d04f6f30c9f52d6ccc1ebfb6c93f687",
	}

	if d := getAddressesInBlockResponse.Result.Data; !reflect.DeepEqual(d, expectedAddressesInTransaction) {
		t.Fatalf("wrong result: %+v", d)
	}
}

func TestLambdaRpcFunctionAddressInRequests(t *testing.T) {
	dbConn, done, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	// Prepate test data
	appearances := []queueItem.Appearance{
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 1},
		{Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b", BlockNumber: 4053179, TransactionIndex: 1},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 2},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 3},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 4},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 5},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionIndex: 6},
	}
	if err = database.InsertAppearanceBatch(context.TODO(), dbConn, appearances); err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse[[]string]{}

	// Adresses in tx

	request = &query.RpcRequest{
		Method: "tb_getAddressesInTransaction",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAddressesInParam{
			{
				BlockNumber:      "4053179",
				TransactionIndex: "1",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}

	// Valid request, appearances found

	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, response)

	if l := len(response.Data); l != 2 {
		t.Fatal("wrong length:", l)
	}

	// Hex block number

	hexResponse := &query.RpcResponse[[]string]{}
	request = &query.RpcRequest{
		Method: "tb_getAddressesInTransaction",
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAddressesInParam{
			{
				BlockNumber:      "0x3DD8BB",
				TransactionIndex: "1",
			},
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}

	// Valid request, appearances found

	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	t.Log(string(output.Payload))
	helpers.UnmarshalLambdaOutput(t, output, hexResponse)

	if !reflect.DeepEqual(response, hexResponse) {
		t.Fatal("wrong response:", hexResponse)
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
	response := &query.RpcResponse[[]database.PublicAppearance]{}
	perPage := uint(10)
	maxIters := len(appearances) / int(perPage)

	// Check basic pagination

	var previousPageId *query.PageId
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
		}
		param := query.RpcGetAppearancesParam{
			Address: appearances[0].Address,
			PerPage: 10,
		}
		if previousPageId != nil {
			if err := param.SetPageId(query.PageIdNoSpecial, previousPageId); err != nil {
				t.Fatal(err)
			}
		}
		err = query.SetParams(
			request,
			[]query.RpcGetAppearancesParam{
				param,
			},
		)
		if err != nil {
			t.Fatal("setting rpc request params:", err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); uint(l) != param.PerPage {
			t.Fatal(i, "-- wrong result count:", l)
		}

		pa := make([]database.PublicAppearance, 0, len(response.Result.Data))
		startIndex := uint(i) * perPage
		endIndex := startIndex + perPage
		for _, item := range appearances[startIndex:endIndex] {
			pa = append(pa, database.PublicAppearance{
				BlockNumber:      strconv.FormatUint(uint64(item.BlockNumber), 10),
				TransactionIndex: strconv.FormatUint(uint64(item.TransactionIndex), 10),
			})
		}
		if r := response.Result.Data; !reflect.DeepEqual(r, pa) {
			t.Fatal(i, "-- wrong result:", r)
		}

		if i == 0 && response.Result.Meta.NextPageId != nil {
			t.Fatal("first page returned NextPageId")
		}

		previousPageId = response.Meta.PreviousPageId
	}

	// Check items

	previousPageId = nil
	var pagingResults = make([]database.PublicAppearance, 0, len(appearances))
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
		}
		param := query.RpcGetAppearancesParam{
			Address: appearances[0].Address,
			PerPage: uint(perPage),
		}
		if err := param.SetPageId(query.PageIdNoSpecial, previousPageId); err != nil {
			t.Fatal(err)
		}
		err = query.SetParams(
			request,
			[]query.RpcGetAppearancesParam{param},
		)
		if err != nil {
			t.Fatal("setting rpc request params:", err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); uint(l) != perPage {
			t.Fatal(i, "-- wrong page len:", l)
		}
		pagingResults = append(pagingResults, response.Result.Data...)

		previousPageId = response.Meta.PreviousPageId
	}

	if l := len(pagingResults); l != len(appearances) {
		t.Fatal("wrong result length", l, "expected", len(appearances))
	}

	for index, pa := range pagingResults {
		app, err := pa.Appearance()
		if err != nil {
			t.Fatal(err)
		}
		if bn := app.BlockNumber; bn != appearances[index].BlockNumber {
			t.Fatal("wrong block number", bn, "expected", appearances[index].BlockNumber)
		}
		if txid := app.TransactionIndex; txid != appearances[index].TransactionIndex {
			t.Fatal("wrong txid", txid, "expected", appearances[index].TransactionIndex)
		}
	}

	// Check pageIds

	previousPageId = nil
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
		}
		param := query.RpcGetAppearancesParam{
			Address: appearances[0].Address,
			PerPage: perPage,
		}
		if err := param.SetPageId(query.PageIdNoSpecial, previousPageId); err != nil {
			t.Fatal(err)
		}
		err = query.SetParams(
			request,
			[]query.RpcGetAppearancesParam{param},
		)
		if err != nil {
			t.Fatal("setting rpc request params:", err)
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)
		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if i == 0 {
			if response.Result.Meta.NextPageId != nil {
				t.Fatal("expected no NextPageId on the first page")
			}
		}
		if i == maxIters-1 {
			if response.Result.Meta.PreviousPageId != nil {
				t.Fatal("expected no PreviousPageId on the last page")
			}
		}
		previousPageId = response.Result.Meta.PreviousPageId
	}

	// Check pageId = "" same as pageId = "latest"

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	param := query.RpcGetAppearancesParam{
		Address: appearances[0].Address,
		PerPage: perPage,
	}
	if err := param.SetPageId(query.PageIdLatest, nil); err != nil {
		t.Fatal(err)
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{param},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	latestApps := response.Result.Data

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	param = query.RpcGetAppearancesParam{
		Address: appearances[0].Address,
		PerPage: perPage,
		PageId:  []byte(""),
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{param},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	if !reflect.DeepEqual(latestApps, response.Result.Data) {
		t.Fatal("wrong results")
	}

	// Check going backwards with "earliest"

	var nextPageId *query.PageId
	pagingResults = make([]database.PublicAppearance, 0, len(appearances))
	for i := 0; i < maxIters; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
		}
		param := query.RpcGetAppearancesParam{
			Address: appearances[0].Address,
			PerPage: perPage,
		}
		var pageIdSpecial query.PageIdSpecial
		var pageId *query.PageId
		if i == 0 {
			pageIdSpecial = query.PageIdEarliest
		} else {
			pageId = nextPageId
		}
		if err := param.SetPageId(pageIdSpecial, pageId); err != nil {
			t.Fatal(err)
		}

		err = query.SetParams(
			request,
			[]query.RpcGetAppearancesParam{param},
		)
		if err != nil {
			t.Fatal("setting rpc request params:", err)
		}

		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result.Data); uint(l) != perPage {
			t.Fatal(i, "-- wrong page len:", l)
		}
		pagingResults = append(pagingResults, response.Result.Data...)

		nextPageId = response.Meta.NextPageId
	}

	if l := len(pagingResults); l != len(appearances) {
		t.Fatal("wrong result length", l, "expected", len(appearances))
	}

	expected := []database.PublicAppearance{
		{BlockNumber: "3001234", TransactionIndex: "10"},
		{BlockNumber: "3001234", TransactionIndex: "9"},
		{BlockNumber: "3001234", TransactionIndex: "8"},
		{BlockNumber: "3001234", TransactionIndex: "7"},
		{BlockNumber: "3001234", TransactionIndex: "6"},
		{BlockNumber: "3001234", TransactionIndex: "5"},
		{BlockNumber: "3001234", TransactionIndex: "4"},
		{BlockNumber: "3001234", TransactionIndex: "3"},
		{BlockNumber: "3001234", TransactionIndex: "2"},
		{BlockNumber: "3001234", TransactionIndex: "1"},
		{BlockNumber: "4053179", TransactionIndex: "20"},
		{BlockNumber: "4053179", TransactionIndex: "19"},
		{BlockNumber: "4053179", TransactionIndex: "18"},
		{BlockNumber: "4053179", TransactionIndex: "17"},
		{BlockNumber: "4053179", TransactionIndex: "16"},
		{BlockNumber: "4053179", TransactionIndex: "15"},
		{BlockNumber: "4053179", TransactionIndex: "14"},
		{BlockNumber: "4053179", TransactionIndex: "13"},
		{BlockNumber: "4053179", TransactionIndex: "12"},
		{BlockNumber: "4053179", TransactionIndex: "11"},
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
	}
	param = query.RpcGetAppearancesParam{
		Address:   appearances[0].Address,
		PerPage:   perPage,
		LastBlock: &raw,
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{param},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	if l := len(response.Result.Data); uint(l) != perPage {
		t.Fatal("wrong page len:", l)
	}
	app, err := response.Result.Data[0].Appearance()
	if err != nil {
		t.Fatal(err)
	}
	if app.BlockNumber != uint32(customLastBlock) {
		t.Fatal("wrong block number")
	}

	// Check max records

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
	}
	param = query.RpcGetAppearancesParam{
		Address: appearances[0].Address,
		PerPage: 1001,
	}
	err = query.SetParams(
		request,
		[]query.RpcGetAppearancesParam{
			param,
		},
	)
	if err != nil {
		t.Fatal("setting rpc request params:", err)
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaProxyError(t, string(output.Payload), query.ErrIncorrectPerPage.Error())
}
