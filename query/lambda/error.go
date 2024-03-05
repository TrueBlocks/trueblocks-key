package main

import (
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
)

type RpcError struct {
	PublicError string
	internal    error
	statusCode  int
}

func NewRpcError(internal error, statusCode int, public string) *RpcError {
	return &RpcError{
		PublicError: public,
		internal:    internal,
		statusCode:  statusCode,
	}
}

func (r *RpcError) Error() string {
	return r.internal.Error()
}

func (r *RpcError) Report(response *events.APIGatewayProxyResponse) {
	log.Println(r.internal)
	response.StatusCode = r.statusCode
	response.Body = strconv.Quote(r.PublicError)
}
