package queue

import (
	"context"
	"encoding/json"
	"fmt"

	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

func (s *SqsQueue) AddBatch(apps []*appearance.Appearance) (err error) {

	entries := make([]types.SendMessageBatchRequestEntry, 0, 10)
	send := func() error {
		msgInput := &sqs.SendMessageBatchInput{
			// Allows batching messages for consumers
			Entries:  entries,
			QueueUrl: &queueUrl,
		}
		_, err := s.awsClient.SendMessageBatch(context.TODO(), msgInput)
		return err
	}

	for _, app := range apps {
		app := app
		if len(entries) == 10 {
			if err := send(); err != nil {
				return err
			}
			entries = make([]types.SendMessageBatchRequestEntry, 0, 10)
		}
		encoded, err := json.Marshal(apps)
		if err != nil {
			return err
		}
		entries = append(entries, types.SendMessageBatchRequestEntry{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Range": {
					DataType:    aws.String("String"),
					StringValue: aws.String(fmt.Sprintf("%d-%d", app.BlockRangeStart, app.BlockRangeEnd)),
				},
			},
			MessageBody: aws.String(string(encoded)),
		})
	}

	if len(entries) > 0 {
		if err = send(); err != nil {
			return
		}
	}

	return
}
