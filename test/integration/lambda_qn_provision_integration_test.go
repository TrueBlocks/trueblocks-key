package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type provisonRequest struct {
	Account *qnaccount.AccountData
	method  string
	path    string
}

func newProvisionRequest(method string, path string) *provisonRequest {
	return &provisonRequest{
		method: method,
		path:   path,
	}
}

func (p *provisonRequest) LambdaPayload() (string, error) {
	body, err := json.Marshal(p.Account)
	if err != nil {
		return "", err
	}
	auth, err := basicAuthValue(awshelper.TestSecret)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
	{
	  "body": %q,
	  "resource": "/{proxy+}",
	  "httpMethod": "%s",
	  "path": "%s",
	  "isBase64Encoded": false,
	  "queryStringParameters": {},
	  "multiValueQueryStringParameters": {},
	  "pathParameters": {
	    "proxy": "%s"
	  },
	  "stageVariables": {
	    "baz": "qux"
	  },
	  "headers": {
	    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	    "Accept-Encoding": "gzip, deflate, sdch",
	    "Accept-Language": "en-US,en;q=0.8",
		"Authorization": "%s",
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
		"Authorization": [
			"%s"
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
	    "path": "%s",
	    "resourcePath": "/{proxy+}",
	    "httpMethod": "%s",
	    "apiId": "1234567890",
	    "protocol": "HTTP/1.1"
	  }
	}
		`, body, p.method, p.path, p.path, auth, auth, p.path, p.method), nil
}

func expectProvisionRequestSuccess(t *testing.T, output *lambda.InvokeOutput) {
	t.Helper()

	// Local API Gateway doesn't handle lambda 401 output correctly and instead of setting
	// the correct status code and headers, it leaves the information in the payload - so we
	// need to parse it.

	provisionResult := make(map[string]any)
	if err := json.Unmarshal(output.Payload, &provisionResult); err != nil {
		t.Fatal("parsing provision payload:", err)
	}
	statusCode, ok := provisionResult["statusCode"]
	if !ok {
		t.Fatal("cannot read provision result status code", fmt.Sprintf("%+v", provisionResult))
	}
	if fmt.Sprint(statusCode) != "200" {
		t.Fatal("provision request failed: status code =", statusCode, fmt.Sprintf("%+v", provisionResult))
	}
}

func expectAuth(t *testing.T, client *lambda.Client, accountData *qnaccount.AccountData, fail bool) {
	t.Helper()

	basicAuth, err := basicAuthValue(awshelper.TestSecret)
	if err != nil {
		t.Fatal("getting basic auth value:", err)
	}
	request := newAuthorizerRequest(map[string]string{
		"Authorization":  basicAuth,
		"x-quicknode-id": accountData.QuicknodeId,
		"x-instance-id":  accountData.EndpointId,
		"x-qn-chain":     accountData.Chain,
		"x-qn-network":   accountData.Network,
	})
	output := helpers.InvokeLambda(t, client, "QnAuthorizer", request)

	if fail {
		helpers.AssertLambdaError(t, string(output.Payload), "Unauthorized")
	} else {
		helpers.AssertLambdaSuccessful(t, output)
	}
}

func TestLambdaQnProvisioning(t *testing.T) {
	done, err := helpers.NewDynamoConnection()
	if err != nil {
		t.Fatal("connecting to dynamo db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	client := helpers.NewLambdaClient(t)
	var request *provisonRequest
	var output *lambda.InvokeOutput

	// Provision

	request = newProvisionRequest("POST", "/provision")
	request.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	expectAuth(t, client, request.Account, false)

	// Provision another endpoint

	request = newProvisionRequest("POST", "/provision")
	request.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id2",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	expectAuth(t, client, request.Account, false)

	// Update

	request = newProvisionRequest("PUT", "/update")
	request.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	expectAuth(t, client, request.Account, false)

	// Deactivate endpoint

	request = newProvisionRequest("DELETE", "/deactivate_endpoint")
	request.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id2",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	expectAuth(t, client, request.Account, true)

	// Deprovision

	request = newProvisionRequest("DELETE", "/deprovision")
	request.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", request)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	expectAuth(t, client, request.Account, true)
}
