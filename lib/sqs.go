package lib

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/golang/glog"
)


type SQSQueue struct {
	sqsClient *sqs.Client
	queueUrl *string
}

type SqsInput struct {
	TransactionHash string
	TransactionMetadata BitCloutTxnMetadata
	PostBody string
}

func NewSQSQueue(client *sqs.Client, queueUrl string) *SQSQueue {
	newSqsQueue := SQSQueue{}
	newSqsQueue.sqsClient = client
	newSqsQueue.queueUrl = &queueUrl
	return &newSqsQueue
}

func (sqsQueue *SQSQueue) SendSQSTxnMessage(txn *MsgBitCloutTxn) {
	var metadata BitCloutTxnMetadata
	if txn.TxnMeta.GetTxnType() == TxnTypeSubmitPost {
		metadata = txn.TxnMeta.(*SubmitPostMetadata)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeLike {
		metadata = txn.TxnMeta.(*LikeMetadata)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeFollow {
		metadata = txn.TxnMeta.(*FollowMetadata)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeBasicTransfer {
		metadata = txn.TxnMeta.(*BasicTransferMetadata)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeCreatorCoin {
		metadata = txn.TxnMeta.(*CreatorCoinMetadataa)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeCreatorCoinTransfer {
		metadata = txn.TxnMeta.(*CreatorCoinTransferMetadataa)
	} else {
		// If we get here then the txn is not a type we're interested in
		return
	}

	var postBody string
	if txn.TxnMeta.GetTxnType() == TxnTypeSubmitPost {
		postBody = string(txn.TxnMeta.(*SubmitPostMetadata).Body)
	} else {
		postBody = ""
	}

	sqsInput := SqsInput {
		TransactionHash: hex.EncodeToString(txn.Hash()[:]),
		TransactionMetadata: metadata,
		PostBody: postBody,
	}

	res, err := json.Marshal(sqsInput)
	if err != nil {
		glog.Errorf("SendSQSTxnMessage: Error marshaling transaction JSON : %v", err)
	}

	sendMessageInput := &sqs.SendMessageInput{
		DelaySeconds: 0,
		MessageBody: aws.String(string(res)),
		QueueUrl:    sqsQueue.queueUrl,
	}
	_, err = sqsQueue.sqsClient.SendMessage(context.TODO(), sendMessageInput)
	if err != nil {
		glog.Error("SendSQSTxnMessage: Error sending sqs message : %v", err)
	}
}