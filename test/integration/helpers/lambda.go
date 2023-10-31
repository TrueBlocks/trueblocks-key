package helpers

import (
	"context"
	"encoding/json"
	"testing"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

type lambdaResponse struct {
	Body string `json:"body"`
}

// NewLambdaClient returns client that allows calling lambda functions locally
func NewLambdaClient(t *testing.T) (client *lambda.Client) {
	t.Helper()
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		t.Fatal("lambda client: loading AWS config failed:", err)
	}

	client = lambda.NewFromConfig(
		cfg,
		func(o *lambda.Options) {
			o.BaseEndpoint = aws.String("http://localhost:3001")
			o.Credentials = credentials.NewStaticCredentialsProvider("fake", "fake", "test")
		},
	)
	return
}

type LambdaPayloadSerializer interface {
	LambdaPayload() (string, error)
}

func InvokeLambda(t *testing.T, client *lambda.Client, function string, payload LambdaPayloadSerializer) *lambda.InvokeOutput {
	t.Helper()
	payloadStr, err := payload.LambdaPayload()
	if err != nil {
		t.Fatal("invoke lambda: serializing payload:", err)
	}
	output, err := client.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(function),
		Payload:      []byte(payloadStr),
	})
	if err != nil {
		t.Fatal("invoke lambda: invoke failed:", err)
	}
	return output
}

func UnmarshalLambdaOutput(t *testing.T, output *lambda.InvokeOutput, target any) {
	t.Helper()
	lr := &lambdaResponse{}
	if err := json.Unmarshal(output.Payload, lr); err != nil {
		t.Fatal("unmarshal lambda output: unmarshal payload:", err)
	}

	if err := json.Unmarshal([]byte(lr.Body), target); err != nil {
		t.Fatal("unmarshal lambda output: unmarshal body:", err)
	}
}

func AssertLambdaSuccessful(t *testing.T, output *lambda.InvokeOutput) {
	t.Helper()
	if output.StatusCode != 200 {
		t.Fatal("assert lambda successful: status code is not 200")
	}

	if output.FunctionError != nil {
		t.Fatal("assert lambda successful:", *output.FunctionError, ":", string(output.Payload))
	}
}

type lambdaError struct {
	ErrorMessage string `json:"errorMessage"`
}

func AssertLambdaError(t *testing.T, payload string, errorStr string) {
	t.Helper()
	lambdaError := &lambdaError{}
	if err := json.Unmarshal([]byte(payload), lambdaError); err != nil {
		t.Fatal("assert lambda error: unmarshal:", err)
	}
	if lambdaError.ErrorMessage != errorStr {
		t.Fatal("assert lamda error: expected", errorStr, "but got", lambdaError.ErrorMessage)
	}
}
