package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

type lambdaResponse struct {
	Body string `json:"body"`
}

// NewLambdaClient returns client that allows calling lambda functions locally
func NewLambdaClient() (client *lambda.Client, err error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return
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

func InvokeLambda(client *lambda.Client, function string, event string) (*lambda.InvokeOutput, error) {
	return client.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(function),
		Payload:      []byte(event),
	})
}

func UnmarshalLambdaOutput(output *lambda.InvokeOutput, target any) error {
	lr := &lambdaResponse{}
	if err := json.Unmarshal(output.Payload, lr); err != nil {
		return fmt.Errorf("json unmarshal payload: %w", err)
	}

	if err := json.Unmarshal([]byte(lr.Body), target); err != nil {
		return fmt.Errorf("json unmarshal body: %w", err)
	}
	return nil
}
