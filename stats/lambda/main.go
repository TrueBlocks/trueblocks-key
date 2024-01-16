package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
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
var dbConn *database.Connection

var ErrInternal = errors.New(http.StatusText(http.StatusInternalServerError))

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	if cnf == nil {
		if err = loadConfig(); err != nil {
			log.Println("loading config:", err)
			err = ErrInternal
			return
		}
	}

	if dbConn == nil {
		if err = setupDbConnection(); err != nil {
			log.Println("database connection:", err)
			err = ErrInternal
			return
		}
	}

	if dynamoClient == nil {
		if err = setupDynamo(); err != nil {
			log.Println("dynamo connection:", err)
			err = ErrInternal
			return
		}
	}

	appCount, err := database.FetchAppearancesCount(ctx, dbConn)
	if err != nil {
		log.Println("fetching appearances count:", err)
		err = ErrInternal
		return
	}

	log.Println("appearances:", appCount)

	chunksCount, err := database.CountChunks(ctx, dbConn)
	if err != nil {
		log.Println("fetching chunks count:", err)
		err = ErrInternal
		return
	}

	log.Println("chunks:", chunksCount)

	dupChunksCount, err := database.FetchDuplicatedChunks(ctx, dbConn)
	if err != nil {
		log.Println("fetching duplicated chunks count:", err)
		err = ErrInternal
		return
	}

	log.Println("duplicated chunks:", dupChunksCount)

	describeOutput, err := dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(cnf.QnProvision.TableName),
	})
	if err != nil {
		log.Println("dynamo describe table:", err)
		err = ErrInternal
		return
	}
	userCount := describeOutput.Table.ItemCount

	log.Println("user count:", userCount)

	body, err := json.Marshal(map[string]any{
		"appearances":      appCount,
		"chunks":           chunksCount,
		"chunksDuplicated": dupChunksCount,
		"users":            userCount,
	})

	response = events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}
	return
}

func loadConfig() (err error) {
	cnf, err = keyConfig.Get("")
	return
}

func setupDynamo() (err error) {
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
	return
}

func setupDbConnection() (err error) {
	var user string
	var password string
	secretId := cnf.Database["default"].AwsSecret
	if secretId != "" {
		log.Println("using Secrets Manager secret as DB password")
		secretValue, err := awshelper.FetchUsernamePasswordSecret(secretId)
		if err != nil {
			return err
		}
		user = secretValue.Username
		password = secretValue.Password
	} else {
		log.Println("using configuration DB password")
		user = cnf.Database["default"].User
		password = cnf.Database["default"].Password
	}

	dbConn = &database.Connection{
		Chain:    "mainnet",
		Host:     cnf.Database["default"].Host,
		Port:     cnf.Database["default"].Port,
		Database: cnf.Database["default"].Database,
		User:     user,
		Password: password,
	}

	log.Println(dbConn.String())

	return dbConn.Connect(context.TODO())
}

func main() {
	defer dbConn.Close(context.TODO())

	lambda.Start(HandleRequest)
}
