package main

import (
	"context"
	"errors"
	"log"

	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/TrueBlocks/trueblocks-key/direct_customers/endpoint"
	keyDynamodb "github.com/TrueBlocks/trueblocks-key/quicknode/keyDynamodb"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var cnf *keyConfig.ConfigFile
var dynamoClient *dynamodb.Client
var errUnauthorized = errors.New("Unauthorized")

func HandleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (result events.APIGatewayCustomAuthorizerResponse, err error) {
	if cnf == nil {
		if cnf, err = keyConfig.Get(""); err != nil {
			log.Println("cannot read configuration")
			return
		}
	}
	if dynamoClient == nil {
		var awsConfig aws.Config
		awsConfig, err = config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Println("error reading config:", err)
			return
		}
		// Create DynamoDB client
		dynamoClient = dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
			if keyDynamodb.ShouldUseLocal() {
				// When running inside sam local (in tests), use local endpoint
				o.BaseEndpoint = aws.String("http://dynamodb:8000")
				o.Credentials = credentials.NewStaticCredentialsProvider("fake", "fake", "test")
			}
		})
	}

	// check account credentials
	endpointId := event.PathParameters["endpointId"]
	if len(endpointId) < 10 {
		log.Println("invalid endpointId: too short")
		err = errUnauthorized
		return
	}

	endpoint, err := endpoint.Find(ctx, dynamoClient, cnf.DirectCustomers.TableName, endpointId)
	if err != nil {
		log.Printf("error finding the endpoint %s: %s\n", endpointId, err)
		err = errUnauthorized
		return
	}

	// TODO: this NEEDS to fetch the plan
	// PrincipalID is something that uniquely identifies the account
	result.PrincipalID = endpoint.ClientId
	result.UsageIdentifierKey = endpoint.ApiKey.Value
	result.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Effect: "allow",
				Action: []string{
					"execute-api:Invoke",
				},
				Resource: []string{
					event.MethodArn,
				},
			},
		},
	}
	return
}

func main() {
	lambda.Start(HandleRequest)
}
