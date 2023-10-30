//go:build integration
// +build integration

package integration_test

import (
	"testing"

	database "trueblocks.io/database/pkg"
	"trueblocks.io/database/pkg/dbtest"
	"trueblocks.io/test/integration/helpers"
)

var eventStr = `
{
    "body": "{ \"jsonrpc\": \"2.0\", \"method\": \"tb_getAppearances\", \"params\": { \"address\": \"0x0000000000000281526004018083600019166000\" } }",
    "resource": "/{proxy+}",
    "path": "/path/to/resource",
    "httpMethod": "POST",
    "isBase64Encoded": true,
    "queryStringParameters": {
        "foo": "bar"
    },
    "multiValueQueryStringParameters": {
        "foo": [
            "bar"
        ]
    },
    "pathParameters": {
        "proxy": "/path/to/resource"
    },
    "stageVariables": {
        "baz": "qux"
    },
    "headers": {
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Encoding": "gzip, deflate, sdch",
        "Accept-Language": "en-US,en;q=0.8",
        "Cache-Control": "max-age=0",
        "CloudFront-Forwarded-Proto": "https",
        "CloudFront-Is-Desktop-Viewer": "true",
        "CloudFront-Is-Mobile-Viewer": "false",
        "CloudFront-Is-SmartTV-Viewer": "false",
        "CloudFront-Is-Tablet-Viewer": "false",
        "CloudFront-Viewer-Country": "US",
        "Host": "1234567890.execute-api.us-east-1.amazonaws.com",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Custom User Agent String",
        "Via": "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)",
        "X-Amz-Cf-Id": "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA==",
        "X-Forwarded-For": "127.0.0.1, 127.0.0.2",
        "X-Forwarded-Port": "443",
        "X-Forwarded-Proto": "https"
    },
    "multiValueHeaders": {
        "Accept": [
            "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
        ],
        "Accept-Encoding": [
            "gzip, deflate, sdch"
        ],
        "Accept-Language": [
            "en-US,en;q=0.8"
        ],
        "Cache-Control": [
            "max-age=0"
        ],
        "CloudFront-Forwarded-Proto": [
            "https"
        ],
        "CloudFront-Is-Desktop-Viewer": [
            "true"
        ],
        "CloudFront-Is-Mobile-Viewer": [
            "false"
        ],
        "CloudFront-Is-SmartTV-Viewer": [
            "false"
        ],
        "CloudFront-Is-Tablet-Viewer": [
            "false"
        ],
        "CloudFront-Viewer-Country": [
            "US"
        ],
        "Host": [
            "0123456789.execute-api.us-east-1.amazonaws.com"
        ],
        "Upgrade-Insecure-Requests": [
            "1"
        ],
        "User-Agent": [
            "Custom User Agent String"
        ],
        "Via": [
            "1.1 08f323deadbeefa7af34d5feb414ce27.cloudfront.net (CloudFront)"
        ],
        "X-Amz-Cf-Id": [
            "cDehVQoZnx43VYQb9j2-nvCh-9z396Uhbp027Y2JvkCPNLmGJHqlaA=="
        ],
        "X-Forwarded-For": [
            "127.0.0.1, 127.0.0.2"
        ],
        "X-Forwarded-Port": [
            "443"
        ],
        "X-Forwarded-Proto": [
            "https"
        ]
    },
    "requestContext": {
        "accountId": "123456789012",
        "resourceId": "123456",
        "stage": "prod",
        "requestId": "c6af9ac6-7b61-11e6-9a41-93e8deadbeef",
        "requestTime": "09/Apr/2015:12:34:56 +0000",
        "requestTimeEpoch": 1428582896000,
        "identity": {
            "cognitoIdentityPoolId": null,
            "accountId": null,
            "cognitoIdentityId": null,
            "caller": null,
            "accessKey": null,
            "sourceIp": "127.0.0.1",
            "cognitoAuthenticationType": null,
            "cognitoAuthenticationProvider": null,
            "userArn": null,
            "userAgent": "Custom User Agent String",
            "user": null
        },
        "path": "/prod/path/to/resource",
        "resourcePath": "/{proxy+}",
        "httpMethod": "POST",
        "apiId": "1234567890",
        "protocol": "HTTP/1.1"
    }
}`

// TODO: this is copy paste
type RpcRequest struct {
	Id     int `json:"id"`
	Params struct {
		Address string `json:"address"`
		Page    int    `json:"page"`
		PerPage int    `json:"perPage"`
	} `json:"params"`
}

type PublicAppearance struct {
	Address       string
	BlockNumber   uint32
	TransactionId uint32
}

type RpcResponse struct {
	JsonRpc string             `json:"jsonrpc"`
	Id      int                `json:"id"`
	Result  []PublicAppearance `json:"result"`
}

func TestValidRequestItemFound(t *testing.T) {
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

	var count int64
	if err = dbConn.Db().Model(&database.Appearance{}).Count(&count).Error; err != nil {
		t.Fatal("count:", err)
	}

	if count != 1 {
		t.Fatal("wrong count:", count)
	}

	client, err := helpers.NewLambdaClient()
	if err != nil {
		t.Fatal("creating lambda client:", err)
	}

	output, err := helpers.InvokeLambda(client, "RpcFunction", eventStr)
	if err != nil {
		t.Fatal("invoke error:", err)
	}
	if output.StatusCode != 200 {
		t.Fatal("status code is not 200")
	}

	if output.FunctionError != nil {
		t.Fatal(*output.FunctionError, ":", string(output.Payload))
	}

	t.Log("payload:", string(output.Payload))

	response := &RpcResponse{}
	if err := helpers.UnmarshalLambdaOutput(output, response); err != nil {
		t.Fatal(err)
	}

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
}
