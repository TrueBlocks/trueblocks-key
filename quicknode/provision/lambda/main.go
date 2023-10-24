package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	qnConfig "trueblocks.io/config/pkg"
	"trueblocks.io/quicknode/secret"
)

var cnf *qnConfig.ConfigFile
var dynamoClient *dynamodb.Client
var ginLambda *ginadapter.GinLambda

func init() {
	if err := loadConfig(); err != nil {
		panic(fmt.Errorf("reading configuration: %w", err))
	}

	secret, err := secret.FetchAuthSecret(cnf.QnProvision.AwsSecret)
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
	if dynamoClient != nil {
		return
	}

	var awsConfig aws.Config
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("error reading config: " + err.Error())
	}
	// Create DynamoDB client
	dynamoClient = dynamodb.NewFromConfig(awsConfig)
}

func loadConfig() (err error) {
	cnf, err = qnConfig.Get("")
	return
}

func main() {
	lambda.Start(HandleRequest)
}
