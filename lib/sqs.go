package lib

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/golang/glog"
)


type SQSQueue struct {
	sqsClient *sqs.Client
	queueUrl *string
}

func NewSQSQueue(client *sqs.Client, queueUrl string) *SQSQueue {
	newSqsQueue := SQSQueue{}
	newSqsQueue.sqsClient = client
	newSqsQueue.queueUrl = &queueUrl
	return &newSqsQueue
}

func (sqsQueue *SQSQueue) SendMessage(message string) {
	sendMessageInput := &sqs.SendMessageInput{
		DelaySeconds: 10,
		MessageAttributes: map[string]types.MessageAttributeValue{
			//"Title": {
			//	DataType:    aws.String("String"),
			//	StringValue: aws.String("The Whistler"),
			//},
			//"Author": {
			//	DataType:    aws.String("String"),
			//	StringValue: aws.String("John Grisham"),
			//},
			//"WeeksOn": {
			//	DataType:    aws.String("Number"),
			//	StringValue: aws.String("6"),
			//},
		},
		MessageBody: aws.String(message),
		QueueUrl:    sqsQueue.queueUrl,
	}
	_, err := sqsQueue.sqsClient.SendMessage(context.TODO(), sendMessageInput);
	if err != nil {
		glog.Error(err)
	}
}