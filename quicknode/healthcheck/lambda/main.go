package main

import (
	"context"
	"log"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var cnf *keyConfig.ConfigFile
var dbConn *database.Connection

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	// For now, we will just try to connect to the index database. If we can connect, then
	// we report success.
	if cnf == nil {
		if err = loadConfig(); err != nil {
			return
		}
	}

	if dbConn == nil {
		if err = setupDbConnection(ctx); err != nil {
			return
		}
	}
	defer dbConn.Close(ctx)

	response = events.APIGatewayProxyResponse{
		Body:       `{ "status": "ok" }`,
		StatusCode: 200,
	}
	return
}

func loadConfig() (err error) {
	cnf, err = keyConfig.Get("")
	return
}

func setupDbConnection(ctx context.Context) (err error) {
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

	return dbConn.Connect(ctx)
}

func main() {
	lambda.Start(HandleRequest)
}
