package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
)

var apiGatewayClient *apigateway.Client

func initApiGateway() (err error) {
	if apiGatewayClient != nil {
		return
	}

	var awsConfig aws.Config
	awsConfig, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("error reading config:", err)
		return
	}

	apiGatewayClient = apigateway.NewFromConfig(awsConfig)
	return nil
}
