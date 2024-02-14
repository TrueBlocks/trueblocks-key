package queue

import (
	"context"
	"encoding/json"
	"fmt"

	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
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

func (s *SqsQueue) Add(itemType queueItem.ItemType, item any) (msgId string, err error) {
	encoded, err := json.Marshal(item)
	if err != nil {
		return
	}

	msgInput := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(itemType)),
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

func (s *SqsQueue) AddAppearanceBatch(items []*queueItem.Appearance) (err error) {
	entries := make([]types.SendMessageBatchRequestEntry, 0, 10)

	for index, item := range items {
		if len(entries) == 10 {
			if err := s.sendBatch(entries); err != nil {
				return err
			}
			entries = make([]types.SendMessageBatchRequestEntry, 0, 10)
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}
		entries = append(entries, types.SendMessageBatchRequestEntry{
			Id: aws.String(fmt.Sprint(index)),
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(string(queueItem.ItemTypeAppearance)),
				},
			},
			MessageBody: aws.String(string(encoded)),
		})
	}

	if len(entries) > 0 {
		if err = s.sendBatch(entries); err != nil {
			return
		}
	}

	return
}

func (s *SqsQueue) AddChunkBatch(items []*queueItem.Chunk) (err error) {
	entries := make([]types.SendMessageBatchRequestEntry, 0, 10)

	for index, item := range items {
		if len(entries) == 10 {
			if err := s.sendBatch(entries); err != nil {
				return err
			}
			entries = make([]types.SendMessageBatchRequestEntry, 0, 10)
		}
		encoded, err := json.Marshal(item)
		if err != nil {
			return err
		}
		entries = append(entries, types.SendMessageBatchRequestEntry{
			Id: aws.String(fmt.Sprint(index)),
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(string(queueItem.ItemTypeChunk)),
				},
			},
			MessageBody: aws.String(string(encoded)),
		})
	}

	if len(entries) > 0 {
		if err = s.sendBatch(entries); err != nil {
			return
		}
	}

	return
}

func (s *SqsQueue) sendBatch(entries []types.SendMessageBatchRequestEntry) error {
	msgInput := &sqs.SendMessageBatchInput{
		// Allows batching messages for consumers
		Entries:  entries,
		QueueUrl: &queueUrl,
	}
	_, err := s.awsClient.SendMessageBatch(context.TODO(), msgInput)
	return err
}
