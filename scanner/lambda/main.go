package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"trueblocks.io/searcher/pkg/query"
)

type Response struct {
	Txs []index.AppearanceRecord `json:"txs"`
}

var runEnv *LambdaRunEnv

func init() {
	runEnv = &LambdaRunEnv{}
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	address := request.QueryStringParameters["address"]
	results := make(chan index.AppearanceRecord, 100)

	go func() {
		if err := query.Find("mainnet", base.HexToAddress(address), runEnv, results); err != nil {
			log.Println("find error: %w", err)
		}
	}()

	var records []index.AppearanceRecord
	for app := range results {
		records = append(records, app)
	}

	log.Println("Found", len(records), "appearances of", address)

	bodyResponse := &Response{
		Txs: records,
	}

	body, err := json.Marshal(bodyResponse)
	if err != nil {
		return
	}

	// TODO: would returning a simple object and rewriting response in API Gateway
	// make this faster?
	response = events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: 200,
	}

	return
}

func main() {
	lambda.Start(HandleRequest)
}
