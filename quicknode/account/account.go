package qnaccount

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Account struct {
	QuicknodeId string `json:"quicknode-id"`
	Plan        string `json:"plan"`
	EndpointId  string `json:"endpoint-id"`
	WssUrl      string `json:"wss-url"`
	HttpUrl     string `json:"http-url"`
	Chain       string `json:"chain"`
	Network     string `json:"network"`
	// Test does not come with request body, it has to be read from
	// request headers
	Test bool `json:"test"`

	dynamoClient    *dynamodb.DynamoDB
	dynamoTableName *string
}

func NewAccount(dynamoSession *dynamodb.DynamoDB, tableName string) *Account {
	return &Account{
		dynamoClient:    dynamoSession,
		dynamoTableName: aws.String(tableName),
	}
}

func (a *Account) DynamoGet() (result *dynamodb.GetItemOutput, err error) {
	result, err = a.dynamoClient.GetItem(&dynamodb.GetItemInput{
		TableName: a.dynamoTableName,
		Key:       a.dynamoKey(),
	})

	return
}

func (a *Account) DynamoPut() (err error) {
	// Note: Example uses non-pointer value: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/dynamo-example-create-table-item.html
	encoded, err := dynamodbattribute.MarshalMap(a)
	if err != nil {
		err = fmt.Errorf("marshal: %w", err)
		return
	}

	_, err = a.dynamoClient.PutItem(&dynamodb.PutItemInput{
		Item:      encoded,
		TableName: a.dynamoTableName,
	})
	if err != nil {
		err = fmt.Errorf("put item: %w", err)
	}
	return
}

func (a *Account) DynamoDelete() (err error) {
	_, err = a.dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       a.dynamoKey(),
		TableName: a.dynamoTableName,
	})
	if err != nil {
		err = fmt.Errorf("delete item: %w", err)
	}
	return
}

func (a *Account) dynamoKey() map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"QuicknodeId": {
			S: aws.String(a.QuicknodeId),
		},
	}
}
