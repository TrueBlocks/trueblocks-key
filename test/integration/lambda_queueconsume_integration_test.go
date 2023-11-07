//go:build integration
// +build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	database "trueblocks.io/database/pkg"
	"trueblocks.io/database/pkg/dbtest"
	"trueblocks.io/queue/consume/pkg/appearance"
	"trueblocks.io/test/integration/helpers"
)

type sqsReceiveEvent string

func NewSqsReceiveEvent(appearance *appearance.Appearance) (s sqsReceiveEvent, err error) {
	encoded, err := json.Marshal(appearance)
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
      "messageAttributes": {},
      "md5OfBody": "7b270e59b47ff90a553787216d55d91d",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:MyQueue",
      "awsRegion": "us-east-1"
    }
  ]
}`,
		strconv.Quote(string(encoded)),
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
	appearance := &appearance.Appearance{
		Address:       "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:   11154177,
		TransactionId: 1,
	}
	appearance.SetAppearanceId()

	client := helpers.NewLambdaClient(t)
	var request helpers.LambdaPayloadSerializer
	var output *lambda.InvokeOutput
	dbAppearances := make([]database.Appearance, 0)

	// Valid request, appearance added

	request, err = NewSqsReceiveEvent(appearance)
	t.Log(request)
	output = helpers.InvokeLambda(t, client, "AppearancesQueueConsume", request)

	helpers.AssertLambdaSuccessful(t, output)

	// Make sure the appearance has been added to the db

	err = dbConn.Db().Where(&database.Appearance{Address: appearance.Address}).Find(&dbAppearances).Error
	if err != nil {
		t.Fatal("fetching appearances from db:", err)
	}

	if l := len(dbAppearances); l != 1 {
		t.Fatal("wrong number of appearances:", l)
	}

	if appId := dbAppearances[0].AppearanceId; appId != appearance.AppearanceId {
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
      "messageAttributes": {},
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

	var count int64
	err = dbConn.Db().Model(&database.Appearance{}).Count(&count).Error
	if err != nil {
		t.Fatal("fetching appearances from db:", err)
	}

	if count != 1 {
		t.Fatal("expected count to be 1, but got", count)
	}
}
