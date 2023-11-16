package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
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

	// Create dummy request only so we can use its BasicAuth method
	// to check if QN Basic Auth is correct
	r, err := http.NewRequest("get", "/", nil)
	if err != nil {
		log.Println("creating dummy request:", err)
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

	found, err := account.Find()
	if err != nil {
		log.Println("finding account:", account.QuicknodeId, err)
		err = errUnauthorized
		return
	}

	if !found {
		log.Println("account not found", account.QuicknodeId)
		err = errUnauthorized
		return
	}

	endpoint := event.Headers["x-instance-id"]
	chain := event.Headers["x-qn-chain"]
	network := event.Headers["x-qn-network"]

	if chain == "" || network == "" {
		log.Println("empty chain or network", chain, network)
		err = errUnauthorized
		return
	}

	if err = qnaccount.ValidateChainNetwork(chain, network, cnf); err != nil {
		log.Println(err, chain, network)
		err = errUnauthorized
		return
	}

	if !account.HasEndpointId(endpoint) {
		log.Println("endpoint not found", endpoint)
		err = errUnauthorized
		return
	}

	if account.ApiKey.Value == "" {
		log.Println("empty Account.ApiKey", account.QuicknodeId)
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
