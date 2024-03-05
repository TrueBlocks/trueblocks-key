package qnaccount

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrAccountNotFound = errors.New("account not found")
var ErrEndpointNotFound = errors.New("endpoint not found")

type Account struct {
	QuicknodeId string   `json:"quicknode-id"`
	Plan        string   `json:"plan"`
	EndpointIds []string `json:"endpoint-id"`
	WssUrl      string   `json:"wss-url"`
	HttpUrl     string   `json:"http-url"`
	Chain       string   `json:"chain"`
	Network     string   `json:"network"`
	// Test does not come with request body, it has to be read from
	// request headers
	Test bool `json:"test"`
	// ApiKey is set by us
	ApiKey ApiKey `json:"api_key"`

	dynamoClient    *dynamodb.Client
	dynamoTableName *string
	dynamoItem      map[string]types.AttributeValue
}

func NewAccount(dynamoClient *dynamodb.Client, tableName string) *Account {
	return &Account{
		dynamoClient:    dynamoClient,
		dynamoTableName: aws.String(tableName),
	}
}

func (a *Account) LoadApiKey(apiGatewayClient *apigateway.Client) error {
	apiKey, err := FindByPlanSlug(apiGatewayClient, a.Plan)
	if err != nil {
		return fmt.Errorf("fetching API key for plan %s: %w", a.Plan, err)
	}
	a.ApiKey = *apiKey
	return nil
}

func (a *Account) Find() (found bool, err error) {
	if err = a.dynamoGet(false); err != nil {
		err = fmt.Errorf("loading account: %w", err)
		return
	}
	if a.dynamoItem == nil {
		log.Println("account not found:", a.QuicknodeId)
		return
	}

	found = true

	if err = attributevalue.UnmarshalMap(a.dynamoItem, a); err != nil {
		err = fmt.Errorf("unmarshalling endpoint ids: %w", err)
	}
	return
}

func (a *Account) Authorize(ad *AccountData) error {
	if a.QuicknodeId != ad.QuicknodeId {
		return ErrAccountNotFound
	}
	if !a.HasEndpointId(ad.EndpointId) {
		return ErrEndpointNotFound
	}
	return nil
}

func (a *Account) dynamoGet(force bool) (err error) {
	if len(a.dynamoItem) > 0 && !force {
		log.Println("account.DynamoGet: using cached DynamoDB item")
		return
	}

	key, err := a.dynamoKey()
	if err != nil {
		err = fmt.Errorf("getting DynamoDB key: %w", err)
		return
	}

	log.Println("account.DynamoGet: table", *a.dynamoTableName, "key:", a.QuicknodeId)

	result, err := a.dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: a.dynamoTableName,
		Key:       key,
	})
	if err != nil {
		err = fmt.Errorf("getting DynamoDB item: %w", err)
		return
	}

	if result == nil {
		return nil
	}

	if result.Item == nil {
		return nil
	}

	a.dynamoItem = result.Item
	log.Println("account.DynamoGet: found account", a.dynamoItem["QuicknodeId"])

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
		err = fmt.Errorf("dynamoDB put: marshal: %w", err)
		return
	}

	_, err = a.dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      encoded,
		TableName: a.dynamoTableName,
	})
	if err != nil {
		err = fmt.Errorf("dynamoDB put: %w", err)
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

func (a *Account) HasEndpointId(endpointId string) bool {
	for _, registeredEndpoint := range a.EndpointIds {
		if registeredEndpoint == endpointId {
			return true
		}
	}
	return false
}

func (a *Account) SetFromAccountData(ad *AccountData) {
	a.QuicknodeId = ad.QuicknodeId
	a.Plan = ad.Plan
	a.WssUrl = ad.WssUrl
	a.HttpUrl = ad.HttpUrl
	a.Chain = ad.Chain
	a.Network = ad.Network
	a.Test = ad.Test
}

func (a *Account) DeactivateEndpoint(endpointId string) (found bool) {
	endpoints := make([]string, 0, len(a.EndpointIds))
	for _, registeredEndpoint := range a.EndpointIds {
		if registeredEndpoint == endpointId {
			found = true
		} else {
			endpoints = append(endpoints, registeredEndpoint)
		}
	}
	a.EndpointIds = endpoints
	return
}

func (a *Account) ActivateEndpoint(endpointId string) {
	a.EndpointIds = append(a.EndpointIds, endpointId)
}
