package integration

import (
	"fmt"
	"net/http"
	"testing"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type authorizerRequest struct {
	headers map[string]string
}

func newAuthorizerRequest(headers map[string]string) *authorizerRequest {
	return &authorizerRequest{
		headers,
	}
}

func (a *authorizerRequest) LambdaPayload() (string, error) {
	var headers string
	for header, value := range a.headers {
		headers += fmt.Sprintf("%q: %q,\n", header, value)
	}

	return fmt.Sprintf(`
{
  "type": "REQUEST",
  "methodArn": "arn:aws:execute-api:us-east-1:123456789012:/test/GET/request",
  "resource": "/request",
  "path": "/request",
  "httpMethod": "GET",
  "headers": {
    "X-AMZ-Date": "20170718T062915Z",
    "Accept": "*/*",
    %s
    "CloudFront-Viewer-Country": "US",
    "CloudFront-Forwarded-Proto": "https",
    "CloudFront-Is-Tablet-Viewer": "false",
    "CloudFront-Is-Mobile-Viewer": "false",
    "User-Agent": "..."
  },
  "queryStringParameters": {
    "QueryString1": "queryValue1"
  },
  "pathParameters": {},
  "stageVariables": {
    "StageVar1": "stageValue1"
  },
  "requestContext": {
    "path": "/request",
    "accountId": "123456789012",
    "resourceId": "05c7jb",
    "stage": "test",
    "requestId": "...",
    "identity": {
      "apiKey": "...",
      "sourceIp": "...",
      "clientCert": {
        "clientCertPem": "CERT_CONTENT",
        "subjectDN": "www.example.com",
        "issuerDN": "Example issuer",
        "serialNumber": "a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1:a1",
        "validity": {
          "notBefore": "May 28 12:30:02 2019 GMT",
          "notAfter": "Aug  5 09:36:04 2021 GMT"
        }
      }
    },
    "resourcePath": "/request",
    "httpMethod": "GET",
    "apiId": ""
  }
}
	`, headers), nil
}

func basicAuthValue(s *awshelper.UsernamePasswordSecret) (string, error) {
	r, err := http.NewRequest("get", "/", nil)
	if err != nil {
		return "", fmt.Errorf("creating dummy request: %w", err)
	}
	r.SetBasicAuth(s.Username, s.Password)
	return r.Header["Authorization"][0], nil
}

func TestLambdaQnAuthorizer(t *testing.T) {
	done, err := helpers.NewDynamoConnection()
	if err != nil {
		t.Fatal("connecting to dynamo db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	client := helpers.NewLambdaClient(t)
	var request *authorizerRequest
	var output *lambda.InvokeOutput
	var basicAuth string

	// Invalid: missing Authorization header

	request = newAuthorizerRequest(map[string]string{})
	output = helpers.InvokeLambda(t, client, "QnAuthorizer", request)
	helpers.AssertLambdaError(t, string(output.Payload), "cannot parse auth header")

	// Invalid: wrong Authorization header

	basicAuth, err = basicAuthValue(&awshelper.UsernamePasswordSecret{
		Username: "wrong",
		Password: "value",
	})
	if err != nil {
		t.Fatal("getting basic auth value:", err)
	}
	request = newAuthorizerRequest(map[string]string{
		"Authorization": basicAuth,
	})
	output = helpers.InvokeLambda(t, client, "QnAuthorizer", request)
	helpers.AssertLambdaError(t, string(output.Payload), "Unauthorized")

	// Invalid: missing x-quicknode-id header

	basicAuth, err = basicAuthValue(awshelper.TestSecret)
	if err != nil {
		t.Fatal("getting basic auth value:", err)
	}
	request = newAuthorizerRequest(map[string]string{
		"Authorization": basicAuth,
	})
	output = helpers.InvokeLambda(t, client, "QnAuthorizer", request)
	helpers.AssertLambdaError(t, string(output.Payload), "Unauthorized")

	// Insert account record into DynamoDB
	provisionRequest := newProvisionRequest("POST", "/provision")
	provisionRequest.Account = &qnaccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", provisionRequest)
	helpers.AssertLambdaSuccessful(t, output)
	expectProvisionRequestSuccess(t, output)

	t.Log("Inserted new account")

	// Invalid: non-existing account

	basicAuth, err = basicAuthValue(awshelper.TestSecret)
	if err != nil {
		t.Fatal("getting basic auth value:", err)
	}
	request = newAuthorizerRequest(map[string]string{
		"Authorization":  basicAuth,
		"x-quicknode-id": "not-found",
	})
	output = helpers.InvokeLambda(t, client, "QnAuthorizer", request)
	helpers.AssertLambdaError(t, string(output.Payload), "Unauthorized")

	// Valid

	basicAuth, err = basicAuthValue(awshelper.TestSecret)
	if err != nil {
		t.Fatal("getting basic auth value:", err)
	}
	request = newAuthorizerRequest(map[string]string{
		"Authorization":  basicAuth,
		"x-quicknode-id": provisionRequest.Account.QuicknodeId,
		"x-instance-id":  provisionRequest.Account.EndpointId,
		"x-qn-chain":     provisionRequest.Account.Chain,
		"x-qn-network":   provisionRequest.Account.Network,
	})
	output = helpers.InvokeLambda(t, client, "QnAuthorizer", request)
	helpers.AssertLambdaSuccessful(t, output)
}
