package helpers

import (
	"context"
	"fmt"

	"github.com/TrueBlocks/trueblocks-key/quicknode/keyDynamodb"
	"github.com/TrueBlocks/trueblocks-key/test/dbtest"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewDynamoConnection() (done func() error, err error) {
	dockerNetwork := dbtest.ContainerNetwork()

	ctx := context.Background()
	dynamoContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "amazon/dynamodb-local",
				Name:         keyDynamodb.LocalDynamoDbSettings.ContainerName,
				ExposedPorts: []string{keyDynamodb.LocalDynamoDbSettings.Port},
				WaitingFor:   wait.ForListeningPort(nat.Port(keyDynamodb.LocalDynamoDbSettings.Port)),
				// Env: map[string]string{},
				Networks: []string{dockerNetwork},
			},
			Started: true,
		},
	)
	if err != nil {
		return
	}
	terminateContainer := func() error {
		return dynamoContainer.Terminate(ctx)
	}

	endpoint, err := dynamoContainer.Endpoint(ctx, "http")
	if err != nil {
		terminateContainer()
		err = fmt.Errorf("cannot get dynamodb endpoint: %w", err)
		return
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		terminateContainer()
		return
	}
	dynamoTableName := "UsersQn"
	dynamoClient := dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
		// When running inside sam local (in tests), use local endpoint
		o.BaseEndpoint = aws.String(endpoint)
		o.Credentials = credentials.NewStaticCredentialsProvider("fake", "fake", "test")
	})
	_, err = dynamoClient.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String(dynamoTableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("QuicknodeId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("QuicknodeId"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1000),
			WriteCapacityUnits: aws.Int64(1000),
		},
	})
	if err != nil {
		terminateContainer()
		err = fmt.Errorf("cannot create dynamodb table: %w", err)
		return
	}

	done = terminateContainer
	return
}
