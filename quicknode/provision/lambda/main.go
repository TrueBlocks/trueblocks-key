package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	awshelper "trueblocks.io/awshelper/pkg"
	qnConfig "trueblocks.io/config/pkg"
	qnDynamodb "trueblocks.io/extract/quicknode/dynamodb"
)

var awsConfig aws.Config
var cnf *qnConfig.ConfigFile
var dynamoClient *dynamodb.Client
var ginLambda *ginadapter.GinLambda

func init() {
	initAwsConfig()

	if err := loadConfig(); err != nil {
		panic(fmt.Errorf("reading configuration: %w", err))
	}

	initApiGateway()

	secret, err := awshelper.FetchUsernamePasswordSecret(cnf.QnProvision.AwsSecret)
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

	ginLambda = ginadapter.New(r)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	initDynamoDb()

	return ginLambda.ProxyWithContext(ctx, request)
}

func initDynamoDb() {
	if dynamoClient != nil {
		return
	}

	// Create DynamoDB client
	dynamoClient = dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
		if qnDynamodb.ShouldUseLocal() {
			// When running inside sam local (in tests), use local endpoint
			o.BaseEndpoint = aws.String("http://dynamodb:8000")
			o.Credentials = credentials.NewStaticCredentialsProvider("fake", "fake", "test")
		}
	})
}

func initAwsConfig() {
	var err error
	awsConfig, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("error reading config: " + err.Error())
	}
}

func loadConfig() (err error) {
	cnf, err = qnConfig.Get("")
	return
}

func main() {
	lambda.Start(HandleRequest)
}
