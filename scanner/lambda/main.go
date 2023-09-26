package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"trueblocks.io/searcher/pkg/query"
	"trueblocks.io/searcher/pkg/query/chunk"
)

type Response struct {
	Txs []chunk.AppearanceRecord `json:"txs"`
}

var runEnv *LambdaRunEnv

func init() {
	runEnv = &LambdaRunEnv{}
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	address := request.QueryStringParameters["address"]
	results := make(chan chunk.AppearanceRecord, 100)

	go func() {
		if err := query.Find("mainnet", address, runEnv, results); err != nil {
			log.Println("find error: %w", err)
		}
	}()

	var records []chunk.AppearanceRecord
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
