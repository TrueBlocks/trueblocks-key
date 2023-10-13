package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	qnConfig "trueblocks.io/config/pkg"
	database "trueblocks.io/database/pkg"
)

var cnf *qnConfig.ConfigFile
var dbConn *database.Connection
var limit = 500

type Response struct {
	Address       string
	BlockNumber   uint32
	TransactionId uint32
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	address := request.QueryStringParameters["address"]

	if cnf == nil {
		if err = loadConfig(); err != nil {
			return
		}
	}

	if dbConn == nil {
		if err = setupDbConnection(); err != nil {
			return
		}
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		limit = int(confLimit)
	}

	items := make([]Response, 0, limit)
	err = dbConn.Db().Where(&database.Appearance{Address: address}).Find(&items).Error
	if err != nil {
		return
	}

	// TODO: would returning non-JSON and rewriting the response in API gateway make it faster?
	body, err := json.Marshal(items)
	if err != nil {
		return
	}
	response = events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}
	return
}

func loadConfig() (err error) {
	cnf, err = qnConfig.Get("")
	return
}

func setupDbConnection() error {
	dbConn = &database.Connection{
		Host:     cnf.Database["default"].Host,
		Port:     cnf.Database["default"].Port,
		Database: cnf.Database["default"].Database,
		User:     cnf.Database["default"].User,
		Password: cnf.Database["default"].Password,
	}
	return dbConn.Connect()
}

func main() {
	lambda.Start(HandleRequest)
}
