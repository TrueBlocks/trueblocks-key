//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	database "trueblocks.io/database/pkg"
	"trueblocks.io/database/pkg/dbtest"
	"trueblocks.io/query/pkg/query"
	"trueblocks.io/test/integration/helpers"
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
	appearance := &database.Appearance{
		Address:       "0x0000000000000281526004018083600019166000",
		BlockNumber:   1,
		TransactionId: 5,
	}
	if err = dbConn.Db().Create(appearance).Error; err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse{}

	// Valid request, appearance found

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: query.RpcRequestParams{
			Address: appearance.Address,
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	t.Logf("result: %+v", response)

	if l := len(response.Result); l != 1 {
		t.Fatal("wrong result count:", l)
	}
	if addr := response.Result[0].Address; addr != appearance.Address {
		t.Fatal("wrong address:", addr)
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
		Params: query.RpcRequestParams{
			Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	helpers.AssertLambdaSuccessful(t, output)
	helpers.UnmarshalLambdaOutput(t, output, response)

	t.Logf("result: %+v", response)

	if l := len(response.Result); l != 0 {
		t.Fatal("wrong result count:", l)
	}

	// Invalid request: no address

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: query.RpcRequestParams{},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "incorrect address")

	// Invalid request: invalid address

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: query.RpcRequestParams{
			Address: "0000000000000281526004018083600019166000",
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "incorrect address")

	// Invalid request: invalid page

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: query.RpcRequestParams{
			Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			Page:    -1,
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: invalid PerPage

	request = &query.RpcRequest{
		Id:     1,
		Method: "tb_getAppearances",
		Params: query.RpcRequestParams{
			Address: "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
			Page:    10,
			PerPage: -1,
		},
	}
	output = helpers.InvokeLambda(t, client, "RpcFunction", request)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "incorrect page or perPage")

	// Invalid request: params out of range

	outOfIntRange := big.NewInt(0)
	// Set outOfIntRange value to max value of int as mentioned here: https://stackoverflow.com/questions/6878590/the-maximum-value-for-an-int-type-in-go
	outOfIntRange.SetString(fmt.Sprint(int((^uint(0))>>1)), 10)
	// Now make it out of range
	outOfIntRange.Add(outOfIntRange, big.NewInt(1))
	rp := rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"test_method\",\"params\":{\"address\":\"0x0000000000000281526004018083600019166000\",\"page\":8,\"perPage\":%s}}"}`, outOfIntRange.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "invalid JSON")

	// Invalid request: insane number as parameter

	insane := big.NewInt(1 << 60)
	rp = rawPayload(fmt.Sprintf(`{"body": "{\"id\":1,\"method\":\"test_method\",\"params\":{\"address\":\"0x0000000000000281526004018083600019166000\",\"page\":8,\"perPage\":%s}}"}`, insane.String()))
	output = helpers.InvokeLambda(t, client, "RpcFunction", rp)

	t.Logf("result: %+v", response)
	helpers.AssertLambdaError(t, string(output.Payload), "incorrect page or perPage")
}

func TestLambdaRpcFunctionPagination(t *testing.T) {
	dbConn, done, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	// Prepate test data
	appearances := []database.Appearance{
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 1},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 2},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 3},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 4},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 5},
		{Address: "0x209c4784ab1e8183cf58ca33cb740efbf3fc18ef", BlockNumber: 4053179, TransactionId: 6},
	}
	if err = dbConn.Db().Create(&appearances).Error; err != nil {
		t.Fatal("inserting test data:", err)
	}

	client := helpers.NewLambdaClient(t)
	var request *query.RpcRequest
	var output *lambda.InvokeOutput
	response := &query.RpcResponse{}

	// Check basic pagination

	for i := 1; i < len(appearances)+1; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: query.RpcRequestParams{
				Address: appearances[0].Address,
				PerPage: 1,
				Page:    i,
			},
		}
		output = helpers.InvokeLambda(t, client, "RpcFunction", request)

		helpers.AssertLambdaSuccessful(t, output)
		helpers.UnmarshalLambdaOutput(t, output, response)

		if l := len(response.Result); l != 1 {
			t.Fatal(i, "-- wrong result count:", l)
		}

		pa := query.PublicAppearance{
			Address:       appearances[i-1].Address,
			BlockNumber:   appearances[i-1].BlockNumber,
			TransactionId: appearances[i-1].TransactionId,
		}
		if r := response.Result; !reflect.DeepEqual(r, []query.PublicAppearance{pa}) {
			t.Fatal(i, "-- wrong result:", r)
		}
	}

	// Check items

	var pagingResults = make([]query.PublicAppearance, 0, len(appearances))
	perPage := 3
	for i := 1; i < (len(appearances)/perPage)+1; i++ {
		request = &query.RpcRequest{
			Id:     1,
			Method: "tb_getAppearances",
			Params: query.RpcRequestParams{
				Address: appearances[0].Address,
				PerPage: perPage,
				Page:    i,
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
		if addr := pa.Address; addr != appearances[index].Address {
			t.Fatal("wrong address", addr, "expected", appearances[index].Address)
		}
		if bn := pa.BlockNumber; bn != appearances[index].BlockNumber {
			t.Fatal("wrong block number", bn, "expected", appearances[index].BlockNumber)
		}
		if txid := pa.TransactionId; txid != appearances[index].TransactionId {
			t.Fatal("wrong txid", txid, "expected", appearances[index].TransactionId)
		}
	}
}
