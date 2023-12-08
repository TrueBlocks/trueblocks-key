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
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var ErrInternal = errors.New(http.StatusText(http.StatusInternalServerError))

var cnf *keyConfig.ConfigFile
var dbConn *database.Connection

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	rpcRequest := &query.RpcRequest{}
	if err = json.Unmarshal([]byte(request.Body), rpcRequest); err != nil {
		response.StatusCode = http.StatusBadRequest
		err = errors.New("invalid JSON")
		return
	}
	if err = rpcRequest.Validate(); err != nil {
		response.StatusCode = http.StatusBadRequest
		return
	}

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

	limit := rpcRequest.Parameters().PerPage
	if limit == 0 {
		// Just in case we forgot to define the limit in configuration
		limit = 20
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		if int(confLimit) < limit {
			limit = int(confLimit)
		}
	}

	offset := rpcRequest.Parameters().Page - 1
	if offset < 0 {
		offset = 0
	}
	offset = offset * limit

	items, err := database.FetchAppearances(ctx, dbConn, rpcRequest.Address(), uint(limit), uint(offset))
	if err != nil {
		log.Println("database query:", err)
		err = ErrInternal
		return
	}

	r := &query.RpcResponse{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result:  items,
	}

	// TODO: would returning non-JSON and rewriting the response in API gateway make it faster?
	body, err := json.Marshal(r)

	if err != nil {
		log.Println("response marshal:", err)
		err = ErrInternal
		return
	}
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
