package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awshelper "trueblocks.io/awshelper/pkg"
	qnConfig "trueblocks.io/config/pkg"
	database "trueblocks.io/database/pkg"
)

var ErrAddressIncorrect = errors.New("incorrect address")

var cnf *qnConfig.ConfigFile
var dbConn *database.Connection

type RpcRequest struct {
	Id     int `json:"id"`
	Params struct {
		Address string `json:"address"`
		Page    int    `json:"page"`
		PerPage int    `json:"perPage"`
	} `json:"params"`
}

func (r *RpcRequest) Address() string {
	return strings.ToLower(r.Params.Address)
}

func (r *RpcRequest) validate() error {
	// Validate address
	if len(r.Params.Address) != 42 {
		return ErrAddressIncorrect
	}
	if r.Params.Address[:2] != "0x" {
		return ErrAddressIncorrect
	}
	if _, err := hex.DecodeString(r.Params.Address[2:]); err != nil {
		return ErrAddressIncorrect
	}

	// Validate pagination
	if r.Params.Page < 0 || r.Params.PerPage < 0 {
		return errors.New("incorrect page or perPage")
	}

	return nil
}

// PublicAppearance has only members that we want to share with
// the outside world
type PublicAppearance struct {
	Address       string
	BlockNumber   uint32
	TransactionId uint32
}

type RpcResponse struct {
	JsonRpc string             `json:"jsonrpc"`
	Id      int                `json:"id"`
	Result  []PublicAppearance `json:"result"`
}

// type RpcResponse

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	rpcRequest := &RpcRequest{}
	if err = json.Unmarshal([]byte(request.Body), rpcRequest); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = "invalid JSON"
		return
	}
	if err = rpcRequest.validate(); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Body = err.Error()
		return
	}

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

	limit := rpcRequest.Params.PerPage
	if limit == 0 {
		// Just in case we forgot to define the limit in configuration
		limit = 500
	}

	if confLimit := cnf.Query.MaxLimit; confLimit > 0 {
		limit = int(confLimit)
	}

	offset := rpcRequest.Params.Page - 1
	if offset < 0 {
		offset = 0
	}

	items := make([]PublicAppearance, 0, limit)
	err = dbConn.Db().Where(&database.Appearance{Address: rpcRequest.Address()}).Limit(limit).Offset(offset).Model(&database.Appearance{}).Find(&items).Error
	if err != nil {
		return
	}

	r := &RpcResponse{
		JsonRpc: "2.0",
		Id:      rpcRequest.Id,
		Result:  items,
	}

	// TODO: would returning non-JSON and rewriting the response in API gateway make it faster?
	body, err := json.Marshal(r)
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
		Host:     cnf.Database["default"].Host,
		Port:     cnf.Database["default"].Port,
		Database: cnf.Database["default"].Database,
		User:     user,
		Password: password,
	}

	log.Println(dbConn.String())

	return dbConn.Connect()
}

func main() {
	lambda.Start(HandleRequest)
}
