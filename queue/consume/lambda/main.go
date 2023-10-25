package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awshelper "trueblocks.io/awshelper/pkg"
	config "trueblocks.io/config/pkg"
	database "trueblocks.io/database/pkg"
	"trueblocks.io/queue/consume/pkg/appearance"
)

var maxBatchSize = 500
var dbConn *database.Connection

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) (err error) {
	if dbConn == nil {
		if err = setupDbConnection(); err != nil {
			return
		}
	}

	recordCount := len(sqsEvent.Records)
	models := make([]appearance.Appearance, 0, recordCount)

	log.Println("Inserting", recordCount, "items")

	for _, record := range sqsEvent.Records {
		app := appearance.Appearance{}
		if err = json.Unmarshal([]byte(record.Body), &app); err != nil {
			return
		}
		app.SetAppearanceId()
		models = append(models, app)
	}

	batchSize := recordCount
	if batchSize > maxBatchSize {
		batchSize = maxBatchSize
	}

	log.Println("Creating database items")
	// With GORM we cannot get a list of failed inserts, so we can't use
	// events.SQSEventResponse.BatchItemFailures (marking only some queue items as failed)
	err = dbConn.Db().CreateInBatches(&models, batchSize).Error

	if err == nil {
		log.Println("Success:", recordCount, "items inserted")
	}

	return
}

func setupDbConnection() (err error) {
	cnf, err := config.Get("")
	if err != nil {
		return err
	}
	if bs := cnf.Sqs.InsertBatchSize; bs > 0 {
		maxBatchSize = int(bs)
	}

	var password string
	secretId := cnf.Database["default"].AwsSecret
	if secretId != "" {
		log.Println("using Secrets Manager secret as DB password")
		password, err = awshelper.FetchSecret(secretId)
		if err != nil {
			return
		}
	} else {
		log.Println("using configuration DB password")
		password = cnf.Database["default"].Password
	}

	dbConn = &database.Connection{
		Host:     cnf.Database["default"].Host,
		Port:     cnf.Database["default"].Port,
		Database: cnf.Database["default"].Database,
		User:     cnf.Database["default"].User,
		Password: password,
	}
	return dbConn.Connect()
}

func main() {
	lambda.Start(HandleRequest)
}
