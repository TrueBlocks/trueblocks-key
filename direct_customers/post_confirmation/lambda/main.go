package main

import (
	"context"
	"fmt"
	"log"

	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/TrueBlocks/trueblocks-key/direct_customers/endpoint"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dynamoClient *dynamodb.Client
var apiGatewayClient *apigateway.Client
var cnf *keyConfig.ConfigFile

func init() {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalln("error reading config:", err)

	}

	dynamoClient = dynamodb.NewFromConfig(awsConfig)
	apiGatewayClient = apigateway.NewFromConfig(awsConfig)

	cnf, err = keyConfig.Get("")
	if err != nil {
		log.Fatalln("loading config:", err)
	}
}

func HandleRequest(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (response events.CognitoEventUserPoolsPostConfirmation, err error) {
	email := event.Request.UserAttributes["email"]

	apiKey, err := findApiKey(cnf.DirectCustomers.DefaultApiKeyName)
	if err != nil {
		log.Println(err)
		return
	}

	e := endpoint.NewEndpoint(email)
	e.ApiKey = *apiKey
	err = e.Save(ctx, dynamoClient, cnf.DirectCustomers.TableName)
	if err != nil {
		log.Println("saving endpoint:", err)
	}
	response = event
	return
}

func findApiKey(apiKeyName string) (foundKey *qnaccount.ApiKey, err error) {
	keysOutput, err := apiGatewayClient.GetApiKeys(context.TODO(), &apigateway.GetApiKeysInput{
		IncludeValues: aws.Bool(true),
	})
	if err != nil {
		err = fmt.Errorf("cannot get api keys: %w", err)
		return
	}
	for _, key := range keysOutput.Items {
		if *key.Name == apiKeyName {
			foundKey = &qnaccount.ApiKey{
				Name:  *key.Name,
				Value: *key.Value,
			}
			break
		}
	}
	if foundKey == nil {
		err = fmt.Errorf("no key found for name %s", apiKeyName)
	}
	return
}

func main() {
	lambda.Start(HandleRequest)
}
