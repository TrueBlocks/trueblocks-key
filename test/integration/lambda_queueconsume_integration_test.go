//go:build integration
// +build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
	"github.com/TrueBlocks/trueblocks-key/test/dbtest"
	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type sqsReceiveEvent string

func NewSqsReceiveEvent(itemType queueItem.ItemType, payload any) (s sqsReceiveEvent, err error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s = sqsReceiveEvent(fmt.Sprintf(`
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "receiptHandle": "MessageReceiptHandle",
      "body": %s,
      "attributes": {
        "ApproximateReceiveCount": "1",
        "SentTimestamp": "1523232000000",
        "SenderId": "123456789012",
        "ApproximateFirstReceiveTimestamp": "1523232000001"
      },
      "messageAttributes": {
		"Type": {
			"DataType": "String",
			"StringValue": "%s"
		}
	  },
      "md5OfBody": "7b270e59b47ff90a553787216d55d91d",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:MyQueue",
      "awsRegion": "us-east-1"
    }
  ]
}`,
		strconv.Quote(string(encoded)),
		itemType,
	))
	return
}
func (s sqsReceiveEvent) LambdaPayload() (string, error) {
	return string(s), nil
}

func TestLambdaQueueConsumeRequests(t *testing.T) {
	var err error
	dbConn, done, err := dbtest.NewTestConnection()
	if err != nil {
		t.Fatal("connecting to test db:", err)
	}
	defer done()
	defer helpers.KillSamOnPanic()

	// Prepate test data
	appearance := &queueItem.Appearance{
		Address:          "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:      11154177,
		TransactionIndex: 1,
	}

	client := helpers.NewLambdaClient(t)
	var request helpers.LambdaPayloadSerializer
	var output *lambda.InvokeOutput
	dbAppearances := make([]database.Appearance, 0)

	// Valid request, appearance added

	request, err = NewSqsReceiveEvent(queueItem.ItemTypeAppearance, appearance)
	t.Log(request)
	output = helpers.InvokeLambda(t, client, "AppearancesQueueConsume", request)

	helpers.AssertLambdaSuccessful(t, output)

	// Make sure the appearance has been added to the db

	dbAppearances, err = database.FetchAppearances(context.TODO(), dbConn, appearance.Address, 1, 0)
	if err != nil {
		t.Fatal("fetching appearances from db:", err)
	}

	if l := len(dbAppearances); l != 1 {
		t.Fatal("wrong number of appearances:", l)
	}

	if appId := dbAppearances[0].BlockNumber; appId != appearance.BlockNumber {
		t.Fatal("mismatched AppearanceId:", appId)
	}

	if appId := dbAppearances[0].TransactionIndex; appId != appearance.TransactionIndex {
		t.Fatal("mismatched AppearanceId:", appId)
	}

	// Invalid request, appearance not added

	request = sqsReceiveEvent(`
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "receiptHandle": "MessageReceiptHandle",
      "body": "this is INVALID",
      "attributes": {
        "ApproximateReceiveCount": "1",
        "SentTimestamp": "1523232000000",
        "SenderId": "123456789012",
        "ApproximateFirstReceiveTimestamp": "1523232000001"
      },
      "messageAttributes": {
		"Type": {
			"DataType": "String",
			"StringValue": "appearance"
		}
	  },
      "md5OfBody": "7b270e59b47ff90a553787216d55d91d",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:MyQueue",
      "awsRegion": "us-east-1"
    }
  ]
}`,
	)
	t.Log(request)
	output = helpers.InvokeLambda(t, client, "AppearancesQueueConsume", request)

	helpers.AssertLambdaError(t, string(output.Payload), "invalid JSON")

	// Number of records in the DB should not change

	var count int
	count, err = dbConn.CountAppearances()
	if err != nil {
		t.Fatal("fetching appearances from db:", err)
	}

	if count != 1 {
		t.Fatal("expected count to be 1, but got", count)
	}

	// Chunks

	chunk := &queueItem.Chunk{
		Cid:    "QmaozNR7DZHQK1ZcU9p7QdrshMvXqWK6gpu5rmrkPdT3L4",
		Range:  "1000-2000",
		Author: "test1",
	}

	// Valid request, chunk added

	request, err = NewSqsReceiveEvent(queueItem.ItemTypeChunk, chunk)
	t.Log(request)
	output = helpers.InvokeLambda(t, client, "AppearancesQueueConsume", request)

	helpers.AssertLambdaSuccessful(t, output)

	count, err = database.CountChunks(context.TODO(), dbConn)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("wrong chunk count:", count)
	}

	// Insert duplicated chunk
	chunk.Cid = "some-other-cid-to-make-it-a-duplicate"
	request, err = NewSqsReceiveEvent(queueItem.ItemTypeChunk, chunk)
	output = helpers.InvokeLambda(t, client, "AppearancesQueueConsume", request)
	helpers.AssertLambdaSuccessful(t, output)

	dups, err := database.FetchDuplicatedChunks(context.TODO(), dbConn)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(dups); l != 1 {
		t.Fatal("wrong duplicated chunk count:", l)
	}
	if dupRange := dups[0]; dupRange != chunk.Range {
		t.Fatalf("wrong range reported: expected %s but got %s", chunk.Range, dupRange)
	}
}
