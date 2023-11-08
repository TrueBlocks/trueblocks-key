package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awshelper "trueblocks.io/awshelper/pkg"
	qnAccount "trueblocks.io/quicknode/account"
	"trueblocks.io/test/integration/helpers"
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

type provisonRequest struct {
	Account *qnAccount.AccountData
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
	provisionRequest.Account = &qnAccount.AccountData{
		QuicknodeId: "test-quicknode-id",
		EndpointId:  "test-endpoint-id",
		Test:        true,
		Plan:        "IntegrationTestPlan",
		Chain:       "ethereum",
		Network:     "mainnet",
	}
	output = helpers.InvokeLambda(t, client, "ApiProvisionFunction", provisionRequest)
	helpers.AssertLambdaSuccessful(t, output)
	// Local API Gateway doesn't handle lambda 401 output correctly and instead of setting
	// the correct status code and headers, it leaves the information in the payload - so we
	// need to parse it.
	provisionResult := make(map[string]any)
	if err = json.Unmarshal(output.Payload, &provisionResult); err != nil {
		t.Fatal("parsing provision payload:", err)
	}
	statusCode, ok := provisionResult["statusCode"]
	if !ok {
		t.Fatal("cannot read provision result status code", fmt.Sprintf("%+v", provisionResult))
	}
	if fmt.Sprint(statusCode) != "200" {
		t.Fatal("provision request failed: status code =", statusCode, fmt.Sprintf("%+v", provisionResult))
	}

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
