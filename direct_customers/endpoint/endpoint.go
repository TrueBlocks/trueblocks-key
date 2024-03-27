package endpoint

import (
	"context"
	"fmt"
	"log"
	"time"

	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	petname "github.com/dustinkirkland/golang-petname"
)

func init() {
	petname.NonDeterministicMode()
}

type Endpoint struct {
	Endpoint string           `json:"endpointId"`
	Email    string           `json:"email"`
	ApiKey   qnaccount.ApiKey `json:"apiKey"`
	// Enabled  bool       `json:"enabled"`
	Created *time.Time `json:"created"`
	// Deleted  *time.Time `json:"deleted"`
}

func NewEndpoint(clientId string) *Endpoint {
	generatedPetname := petname.Generate(3, "-")

	created := time.Now()
	log.Println("new endpoint:", generatedPetname)

	return &Endpoint{
		Endpoint: generatedPetname,
		Email:    clientId,
		Created:  &created,
	}
}

func Find(ctx context.Context, dynamoClient *dynamodb.Client, tableName string, endpointId string) (e *Endpoint, err error) {
	key, err := encodeEndpointId(endpointId)
	if err != nil {
		return
	}

	record, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return
	}

	e = &Endpoint{}
	if err = attributevalue.UnmarshalMap(record.Item, e); err != nil {
		err = fmt.Errorf("unmarshalling endpoint: %w", err)
	}
	return
}

func (e *Endpoint) Save(ctx context.Context, dynamoClient *dynamodb.Client, tableName string) (err error) {
	key, err := encodeEndpointId(e.Endpoint)
	if err != nil {
		return
	}

	_, err = dynamoClient.UpdateItem(
		ctx,
		&dynamodb.UpdateItemInput{
			TableName:        aws.String(tableName),
			Key:              key,
			UpdateExpression: aws.String("SET Email = :email, apiKey = :apiKey"),
			ExpressionAttributeValues: map[string]dynamodbTypes.AttributeValue{
				":email": &dynamodbTypes.AttributeValueMemberS{
					Value: e.Email,
				},
				":apiKey": &dynamodbTypes.AttributeValueMemberM{
					Value: map[string]dynamodbTypes.AttributeValue{
						"Name": &dynamodbTypes.AttributeValueMemberS{
							Value: e.ApiKey.Name,
						},
						"Value": &dynamodbTypes.AttributeValueMemberS{
							Value: e.ApiKey.Value,
						},
					},
				},
			},
		},
	)

	return
}

func encodeEndpointId(endpointId string) (key map[string]types.AttributeValue, err error) {
	encodedEndpointId, err := attributevalue.Marshal(endpointId)
	if err != nil {
		return
	}
	key = map[string]types.AttributeValue{
		"EndpointId": encodedEndpointId,
	}
	return
}
