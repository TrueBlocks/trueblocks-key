package main

import "github.com/aws/aws-sdk-go-v2/service/apigateway"

var apiGatewayClient *apigateway.Client

func initApiGateway() {
	if apiGatewayClient != nil {
		return
	}
	apiGatewayClient = apigateway.New(apigateway.Options{
		AppID: "qn-provision-lambda",
	})
}
