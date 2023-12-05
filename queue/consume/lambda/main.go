package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var maxBatchSize = 500
var dbConn *database.Connection

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (err error) {
	if dbConn == nil {
		if err = setupDbConnection(ctx); err != nil {
			return
		}
	}
	defer dbConn.Close(ctx)

	recordCount := len(sqsEvent.Records)
	models := make([]appearance.Appearance, 0, recordCount)

	log.Println("Inserting", recordCount, "items")

	for _, record := range sqsEvent.Records {
		app := appearance.Appearance{}
		if err = json.Unmarshal([]byte(record.Body), &app); err != nil {
			log.Println("unmarshal JSON:", err)
			err = errors.New("invalid JSON")
			return
		}
		models = append(models, app)
	}

	// batchSize := recordCount
	// if batchSize > maxBatchSize {
	// 	batchSize = maxBatchSize
	// }

	log.Println("Creating database items")

	err = database.InsertAppearanceBatch(ctx, dbConn, models)
	if err == nil {
		log.Println("Success:", recordCount, "items inserted")
	}

	return
}

func setupDbConnection(ctx context.Context) (err error) {
	cnf, err := config.Get("")
	if err != nil {
		return err
	}
	if bs := cnf.Sqs.InsertBatchSize; bs > 0 {
		maxBatchSize = int(bs)
	}

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
	err = dbConn.Connect(ctx)
	return
}

func main() {
	lambda.Start(HandleRequest)
}
