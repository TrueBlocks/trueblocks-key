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
	petname "github.com/dustinkirkland/golang-petname"
)

func init() {
	petname.NonDeterministicMode()
}

type Endpoint struct {
	Endpoint string           `json:"endpointId"`
	ClientId string           `json:"clientId"`
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
		ClientId: clientId,
		Created:  &created,
	}
}

func Find(ctx context.Context, dynamoClient *dynamodb.Client, tableName string, endpointId string) (e *Endpoint, err error) {
	encodedEndpointId, err := attributevalue.Marshal(endpointId)
	if err != nil {
		return
	}
	key := map[string]types.AttributeValue{
		"EndpointId": encodedEndpointId,
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
