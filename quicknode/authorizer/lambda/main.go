package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	qnConfig "trueblocks.io/config/pkg"
	qnaccount "trueblocks.io/quicknode/account"
	"trueblocks.io/quicknode/authorizer/pkg/plan"
	"trueblocks.io/quicknode/secret"
)

var cnf *qnConfig.ConfigFile
var svc *dynamodb.DynamoDB
var apiGatewayClient *apigateway.Client
var errUnauthorized = errors.New("Unauthorized")

func HandleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (result events.APIGatewayCustomAuthorizerResponse, err error) {
	if cnf == nil {
		if cnf, err = qnConfig.Get(""); err != nil {
			log.Println("cannot read configuration")
			return
		}
	}
	if svc == nil {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		// Create DynamoDB client
		svc = dynamodb.New(sess)
	}
	if apiGatewayClient == nil {
		apiGatewayClient = apigateway.New(apigateway.Options{
			AppID: "qn-authorizer",
		})
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

	secret, err := secret.FetchAuthSecret(cnf.QnProvision.AwsSecret)
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

	account := qnaccount.NewAccount(svc, cnf.QnProvision.TableName)
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

	planItem := getResult.Item["Plan"]
	if planItem == nil {
		log.Println("plan is nil", account.QuicknodeId)
		err = errUnauthorized
		return
	}

	planSlug := planItem.GoString()
	apiKey, err := plan.FindBySlug(apiGatewayClient, planSlug)
	if err != nil {
		log.Println(err)
		err = errUnauthorized
		return
	}

	// PrincipalID is something that uniquely identifies the account
	result.PrincipalID = account.QuicknodeId
	result.UsageIdentifierKey = apiKey
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
