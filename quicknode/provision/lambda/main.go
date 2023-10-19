package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	qnConfig "trueblocks.io/config/pkg"
)

var cnf *qnConfig.ConfigFile
var svc *dynamodb.DynamoDB
var ginLambda *ginadapter.GinLambda
var dynamoTableName *string

func init() {
	if err := loadConfig(); err != nil {
		panic(fmt.Errorf("reading configuration: %w", err))
	}

	dynamoTableName = aws.String(cnf.QnProvision.TableName)
	secret, err := fetchAuthSecret()
	if err != nil {
		panic(err)
	}
	r := gin.Default()
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		secret.Username: secret.Password,
	}))

	authorized.POST("/provision", HandleProvision)
	authorized.PUT("/update", HandleUpdate)
	authorized.DELETE("/deactivate_endpoint", HandleDeactivateEndpoint)
	authorized.DELETE("/deprovision", HandleDeprovision)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	initDynamoDb()

	return ginLambda.ProxyWithContext(ctx, request)
}

func initDynamoDb() {
	if svc == nil {
		return
	}

	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc = dynamodb.New(sess)
}

func loadConfig() (err error) {
	cnf, err = qnConfig.Get("")
	return
}

func main() {
	lambda.Start(HandleRequest)
}
