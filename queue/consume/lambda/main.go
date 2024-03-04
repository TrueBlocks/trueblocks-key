package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var maxBatchSize = 500
var dbConn *database.Connection

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (err error) {
	if err = setupDbConnection(ctx); err != nil {
		return
	}
	defer dbConn.Close(context.TODO())

	recordCount := len(sqsEvent.Records)
	appearances := make([]queueItem.Appearance, 0, recordCount)
	chunks := make([]queueItem.Chunk, 0, recordCount)

	log.Println("Inserting", recordCount, "items")

	for _, record := range sqsEvent.Records {
		var recordType string
		rawType := record.MessageAttributes["Type"].StringValue
		if rawType == nil {
			recordType = ""
		} else {
			recordType = *rawType
		}
		switch recordType {
		case string(queueItem.ItemTypeAppearance):
			item := queueItem.Appearance{}
			if err = json.Unmarshal([]byte(record.Body), &item); err != nil {
				log.Println("unmarshal appearance JSON:", err)
				err = errors.New("invalid JSON")
				return
			}
			appearances = append(appearances, item)
		case string(queueItem.ItemTypeChunk):
			item := queueItem.Chunk{}
			if err = json.Unmarshal([]byte(record.Body), &item); err != nil {
				log.Println("unmarshal chunk JSON:", err)
				err = errors.New("invalid JSON")
				return
			}
			chunks = append(chunks, item)
		default:
			err = fmt.Errorf("unsupported message type: %s", recordType)
		}
	}

	log.Println("Creating database items")

	if len(appearances) > 0 {
		log.Println("inserting appearances")

		err = database.InsertAppearanceBatch(ctx, dbConn, appearances)
		if err == nil {
			log.Println("Success:", recordCount, "items inserted")
		}
	}

	if len(chunks) > 0 {
		log.Println("inserting chunks")

		err = database.InsertChunkBatch(ctx, dbConn, chunks)
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
