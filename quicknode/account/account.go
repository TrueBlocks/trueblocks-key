package qnaccount

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	// ApiKey is set by us
	ApiKey ApiKey `json:"api_key"`

	dynamoClient    *dynamodb.Client
	dynamoTableName *string
}

func NewAccount(dynamoClient *dynamodb.Client, tableName string) *Account {
	return &Account{
		dynamoClient:    dynamoClient,
		dynamoTableName: aws.String(tableName),
	}
}

func (a *Account) DynamoGet() (item map[string]types.AttributeValue, err error) {
	key, err := a.dynamoKey()
	if err != nil {
		return
	}
	result, err := a.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: a.dynamoTableName,
		Key:       key,
	})
	if err != nil {
		return
	}

	if result == nil {
		return nil, nil
	}

	if result.Item == nil {
		return nil, nil
	}

	return
}

func (a *Account) DynamoPut() (err error) {
	if a.ApiKey.Value == "" {
		err = errors.New("cannot put with empty ApiKey.Value")
		return
	}

	// Note: Example uses non-pointer value: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/dynamo-example-create-table-item.html
	encoded, err := attributevalue.MarshalMap(a)
	if err != nil {
		err = fmt.Errorf("marshal: %w", err)
		return
	}

	_, err = a.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      encoded,
		TableName: a.dynamoTableName,
	})
	if err != nil {
		err = fmt.Errorf("put item: %w", err)
	}
	return
}

func (a *Account) DynamoDelete() (err error) {
	key, err := a.dynamoKey()
	if err != nil {
		return
	}
	_, err = a.dynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key:       key,
		TableName: a.dynamoTableName,
	})
	if err != nil {
		err = fmt.Errorf("delete item: %w", err)
	}
	return
}

func (a *Account) dynamoKey() (map[string]types.AttributeValue, error) {
	id, err := attributevalue.Marshal(a.QuicknodeId)
	if err != nil {
		return nil, err
	}
	return map[string]types.AttributeValue{
		"QuicknodeId": id,
	}, nil
}
