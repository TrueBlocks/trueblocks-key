package integration

import (
	"context"
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/TrueBlocks/trueblocks-key/test/dbtest"
	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type statsRequest struct{}

func (s *statsRequest) LambdaPayload() (string, error) {
	return `
	{
	  "resource": "/{proxy+}",
	  "httpMethod": "GET",
	  "path": "/stats",
	  "isBase64Encoded": false,
	  "queryStringParameters": {},
	  "multiValueQueryStringParameters": {},
	  "pathParameters": {
	    "proxy": "/stats"
	  },
	  "stageVariables": {
	    "baz": "qux"
	  },
	  "headers": {
	    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	    "Accept-Encoding": "gzip, deflate, sdch",
	    "Accept-Language": "en-US,en;q=0.8",
		"Content-Type": "application/json",
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
	    "path": "/stats",
	    "resourcePath": "/{proxy+}",
	    "httpMethod": "GET",
	    "apiId": "1234567890",
	    "protocol": "HTTP/1.1"
	  }
	}
		`, nil
}

func TestStatsApi(t *testing.T) {
	dbConn, dbDone, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer dbDone()
	dynamoDone, err := helpers.NewDynamoConnection()
	if err != nil {
		t.Fatal("connecting to dynamo db:", err)
	}
	defer dynamoDone()
	defer helpers.KillSamOnPanic()

	client := helpers.NewLambdaClient(t)
	var output *lambda.InvokeOutput

	// Insert test account

	provisionRequest := newProvisionRequest("POST", "/provision")
	provisionRequest.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "QnProvisionFunction", provisionRequest)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	t.Log("Inserted new account")

	// Insert test appearance

	address := "0x0000000000000281526004018083600019166000"
	appearance := &database.Appearance{
		BlockNumber:   1,
		TransactionId: 5,
	}
	if err = appearance.Insert(context.TODO(), dbConn, address); err != nil {
		t.Fatal("inserting test data:", err)
	}

	// Retrieve stats

	var response struct {
		Appearances int `json:"appearances"`
		Users       int `json:"users"`
	}
	output = helpers.InvokeLambda(t, client, "StatsFunction", &statsRequest{})
	helpers.UnmarshalLambdaOutput(t, output, &response)

	t.Log(response)

	if c := response.Appearances; c != 1 {
		t.Fatal("wrong appearance count:", c)
	}
	if c := response.Users; c != 1 {
		t.Fatal("wrong users count:", c)
	}
}
