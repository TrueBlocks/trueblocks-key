package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	awshelper "github.com/TrueBlocks/trueblocks-key/awshelper/pkg"
	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/query/pkg/query"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const defaultLimit = 100

var ErrInternal = errors.New(http.StatusText(http.StatusInternalServerError))

var cnf *keyConfig.ConfigFile
var dbConn *database.Connection

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	rpcRequest := &query.RpcRequest{}
	if err = json.Unmarshal([]byte(request.Body), rpcRequest); err != nil {
		// response.StatusCode = http.StatusBadRequest
		// response.Body = strconv.Quote("invalid JSON")
		NewRpcError(err, http.StatusBadRequest, "invalid JSON").Report(&response)
		err = nil
		return
	}
	// if err = rpcRequest.Validate(); err != nil {
	// 	response.StatusCode = http.StatusBadRequest
	// 	response.Body = strconv.Quote(err.Error())
	// 	err = nil
	// 	return
	// }

	if cnf == nil {
		if err = loadConfig(); err != nil {
			log.Println("loading config:", err)
			err = ErrInternal
			return
		}
	}

	// When working with RDS Proxy we don't "cache" the connection
	// between lambda invocations, so we need to recreate it each time
	if err = setupDbConnection(ctx); err != nil {
		log.Println("database connection:", err)
		err = ErrInternal
		return
	}

	var r any
	switch rpcRequest.Method {
	case query.MethodGetAppearances:
		r, err = handleGetAppearances(ctx, rpcRequest)
	case query.MethodGetBounds:
		r, err = handleBounds(ctx, rpcRequest)
	case query.MethodLastIndexedBlock:
		r, err = handleLastIndexedBlock(ctx, rpcRequest)
	case query.MethodGetAddressesInTx:
		r, err = handleGetAddressesInTx(ctx, rpcRequest)
	default:
		err = fmt.Errorf("unsupported method: %s", rpcRequest.Method)
		err = NewRpcError(err, http.StatusBadRequest, err.Error())
	}
	// When working with RDS Proxy we have to close db connection as soon
	// as possible
	if closeErr := dbConn.Close(ctx); closeErr != nil {
		log.Println("error while closing db connection:", closeErr)
	}

	if err != nil {
		if rpcErr, ok := err.(*RpcError); ok {
			rpcErr.Report(&response)
			err = nil
		} else {
			log.Println(err)
		}

		return
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

func unmarshallAndValidateRequest[T query.Validator](encoded string, response events.APIGatewayProxyResponse) (rpcRequest T, err error) {
	// rpcRequest = new(T)
	if err = json.Unmarshal([]byte(encoded), rpcRequest); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = strconv.Quote("invalid JSON")
		// err = nil
		// return
	}
	if err = rpcRequest.Validate(); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = strconv.Quote(err.Error())
		// err = nil
		// return
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
