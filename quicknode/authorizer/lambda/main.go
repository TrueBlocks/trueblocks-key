package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awshelper "trueblocks.io/awshelper/pkg"
	qnConfig "trueblocks.io/config/pkg"
	qnaccount "trueblocks.io/quicknode/account"
)

var cnf *qnConfig.ConfigFile
var dynamoClient *dynamodb.Client
var errUnauthorized = errors.New("Unauthorized")

func HandleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (result events.APIGatewayCustomAuthorizerResponse, err error) {
	if cnf == nil {
		if cnf, err = qnConfig.Get(""); err != nil {
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
		dynamoClient = dynamodb.NewFromConfig(awsConfig)
	}

	// Create dummy request only so we can use it's BasicAuth method
	// to check if QN Basic Auth is correct
	r, err := http.NewRequest("get", "/", nil)
	if err != nil {
		log.Println("parsing basic auth header:", err)
		return
	}
	r.Header.Add("Authorization", event.Headers["Authorization"])
	username, password, ok := r.BasicAuth()
	if !ok {
		err = errors.New("cannot parse auth header")
		log.Println(err.Error())
		return
	}

	secret, err := awshelper.FetchUsernamePasswordSecret(cnf.QnProvision.AwsSecret)
	if err != nil {
		log.Println("cannot read secret")
		return
	}

	if secret.Username != username || secret.Password != password {
		log.Println("QN auth header invalid")
		err = errUnauthorized
		return
	}

	// QN Basic Auth is correct, we can now check account credentials

	account := qnaccount.NewAccount(dynamoClient, cnf.QnProvision.TableName)
	account.QuicknodeId = event.Headers["x-quicknode-id"]

	if account.QuicknodeId == "" {
		log.Println("empty QuicknodeId")
		err = errUnauthorized
		return
	}

	getResult, err := account.DynamoGet()
	if err != nil {
		log.Println("cannot get from account DynamoDB:", account.QuicknodeId)
		err = errUnauthorized
		return
	}
	if getResult == nil {
		log.Println("no such account:", account.QuicknodeId)
		err = errUnauthorized
		return
	}

	apiKeyAttr, ok := getResult["ApiKey"]
	if !ok || apiKeyAttr == nil {
		log.Println("empty Account.ApiKey", account.QuicknodeId)
		err = errUnauthorized
		return
	}
	if err = attributevalue.Unmarshal(apiKeyAttr, account.ApiKey); err != nil {
		log.Println("unmarshal Account.ApiKey:", err)
		err = errUnauthorized
		return
	}

	// PrincipalID is something that uniquely identifies the account
	result.PrincipalID = account.QuicknodeId
	result.UsageIdentifierKey = account.ApiKey.Value
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
