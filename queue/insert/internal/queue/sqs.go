package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	config "trueblocks.io/config/pkg"
	"trueblocks.io/queue/consume/pkg/appearance"
)

var queueUrl string

type SqsQueue struct {
	awsClient *sqs.Client
	queueName string
}

func NewSqsQueue(awsClient *sqs.Client, keyConfig *config.ConfigFile) *SqsQueue {
	return &SqsQueue{
		awsClient: awsClient,
		queueName: keyConfig.Sqs.QueueName,
	}
}

func (s *SqsQueue) Init() error {
	if queueUrl == "" {
		getUrlResult, err := s.awsClient.GetQueueUrl(
			context.TODO(),
			&sqs.GetQueueUrlInput{QueueName: &s.queueName},
		)
		if err != nil {
			return err
		}

		queueUrl = *getUrlResult.QueueUrl
	}
	return nil
}

func (s *SqsQueue) Add(app *appearance.Appearance) (msgId string, err error) {
	encoded, err := json.Marshal(app)
	if err != nil {
		return
	}

	msgInput := &sqs.SendMessageInput{
		// Allows batching messages for consumers
		DelaySeconds: 10,
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Range": {
				DataType:    aws.String("String"),
				StringValue: aws.String(fmt.Sprintf("%d-%d", app.BlockRangeStart, app.BlockRangeEnd)),
			},
		},
		MessageBody: aws.String(string(encoded)),
		QueueUrl:    &queueUrl,
	}
	output, err := s.awsClient.SendMessage(context.TODO(), msgInput)
	if err != nil {
		return
	}
	msgId = *output.MessageId
	return
}
