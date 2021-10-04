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
	TransactionType string
	TransactionHash string
	NotificationData BitCloutNotification
}

type BitCloutNotification interface {

}

func NewSQSQueue(client *sqs.Client, queueUrl string) *SQSQueue {
	newSqsQueue := SQSQueue{}
	newSqsQueue.sqsClient = client
	newSqsQueue.queueUrl = &queueUrl
	return &newSqsQueue
}

type SubmitPostNotification struct {
	AffectedPublicKeys []*AffectedPublicKey
	PostHashToModify string
	ParentStakeId string
	Body string
	CreatorBasisPoints uint64
	StakeMultipleBasisPoints uint64
	TimestampNanos uint64
	IsHidden bool
}

type LikeNotification struct {
	AffectedPublicKeys []*AffectedPublicKey
	LikedPostHashHex string
	IsUnlike bool
}

type FollowNotification struct {
	AffectedPublicKeys []*AffectedPublicKey
	FollowedPublicKey string
	IsUnfollow bool
}

type BasicTransferNotification struct {

}

type CreatorCoinNotification struct {
	AffectedPublicKeys []*AffectedPublicKey
	ProfilePublicKey string
	OperationType CreatorCoinOperationType
	BitCloutToSellNanos    uint64
	CreatorCoinToSellNanos uint64
	BitCloutToAddNanos     uint64
	MinBitCloutExpectedNanos    uint64
	MinCreatorCoinExpectedNanos uint64
}

type CreatorCoinTransferNotification struct {
	AffectedPublicKeys []*AffectedPublicKey
	ProfilePublicKey string
	CreatorCoinToTransferNanos uint64
	ReceiverPublicKey          string
}


func (sqsQueue *SQSQueue) SendSQSTxnMessage(mempoolTxn *MempoolTx) {
	txn := mempoolTxn.Tx
	var notificationData BitCloutNotification
	if txn.TxnMeta.GetTxnType() == TxnTypeSubmitPost {
		notificationData = makeSubmitPostNotification(mempoolTxn)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeLike {
		notificationData = makeLikeNotification(mempoolTxn)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeFollow {
		notificationData = makeFollowNotification(mempoolTxn)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeBasicTransfer {
		notificationData = makeBasicTransferNotification(mempoolTxn)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeCreatorCoin {
		notificationData = makeCreatorCoinNotification(mempoolTxn)
	} else if txn.TxnMeta.GetTxnType() == TxnTypeCreatorCoinTransfer {
		notificationData = makeCreatorCoinTransferNotification(mempoolTxn)
	} else {
		// If we get here then the txn is not a type we're interested in
		return
	}

	sqsInput := SqsInput {
		TransactionType: txn.TxnMeta.GetTxnType().String(),
		TransactionHash: hex.EncodeToString(txn.Hash()[:]),
		NotificationData: notificationData,
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

func makeSubmitPostNotification(mempoolTxn *MempoolTx) (*SubmitPostNotification){
	metadata := mempoolTxn.Tx.TxnMeta.(*SubmitPostMetadata)
	affectedPublicKeys := mempoolTxn.TxMeta.AffectedPublicKeys
	return &SubmitPostNotification {
		AffectedPublicKeys: affectedPublicKeys,
		PostHashToModify: hex.EncodeToString(metadata.PostHashToModify),
		ParentStakeId: hex.EncodeToString(metadata.ParentStakeID),
		Body: string(metadata.Body),
		CreatorBasisPoints: metadata.CreatorBasisPoints,
		StakeMultipleBasisPoints: metadata.StakeMultipleBasisPoints,
		TimestampNanos: metadata.TimestampNanos,
		IsHidden: metadata.IsHidden,
	}
}

func makeLikeNotification(mempoolTxn *MempoolTx) (*LikeNotification) {
	metadata := mempoolTxn.Tx.TxnMeta.(*LikeMetadata)
	affectedPublicKeys := mempoolTxn.TxMeta.AffectedPublicKeys
	return &LikeNotification{
		AffectedPublicKeys: affectedPublicKeys,
		LikedPostHashHex: hex.EncodeToString([]byte(metadata.LikedPostHash[:])),
		IsUnlike: metadata.IsUnlike,
	}
}

func makeFollowNotification(mempoolTxn *MempoolTx) (*FollowNotification) {
	metadata := mempoolTxn.Tx.TxnMeta.(*FollowMetadata)
	affectedPublicKeys := mempoolTxn.TxMeta.AffectedPublicKeys
	return &FollowNotification{
		AffectedPublicKeys: affectedPublicKeys,
		FollowedPublicKey: hex.EncodeToString(metadata.FollowedPublicKey),
		IsUnfollow: metadata.IsUnfollow,
	}
}

func makeBasicTransferNotification(mempoolTxn *MempoolTx) (*BasicTransferNotification) {
	return &BasicTransferNotification{}
}

func makeCreatorCoinNotification(mempoolTxn *MempoolTx) (*CreatorCoinNotification) {
	metadata := mempoolTxn.Tx.TxnMeta.(*CreatorCoinMetadataa)
	affectedPublicKeys := mempoolTxn.TxMeta.AffectedPublicKeys
	return &CreatorCoinNotification {
		AffectedPublicKeys: affectedPublicKeys,
		ProfilePublicKey: hex.EncodeToString(metadata.ProfilePublicKey),
		OperationType: metadata.OperationType,
		BitCloutToSellNanos: metadata.BitCloutToSellNanos,
		CreatorCoinToSellNanos: metadata.CreatorCoinToSellNanos,
		BitCloutToAddNanos: metadata.BitCloutToAddNanos,
		MinBitCloutExpectedNanos: metadata.MinBitCloutExpectedNanos,
		MinCreatorCoinExpectedNanos: metadata.MinCreatorCoinExpectedNanos,
	}
}

func makeCreatorCoinTransferNotification(mempoolTxn *MempoolTx) (*CreatorCoinTransferNotification) {
	metadata := mempoolTxn.Tx.TxnMeta.(*CreatorCoinTransferMetadataa)
	affectedPublicKeys := mempoolTxn.TxMeta.AffectedPublicKeys
	return &CreatorCoinTransferNotification {
		AffectedPublicKeys: affectedPublicKeys,
		ProfilePublicKey: hex.EncodeToString(metadata.ProfilePublicKey),
		CreatorCoinToTransferNanos: metadata.CreatorCoinToTransferNanos,
		ReceiverPublicKey: hex.EncodeToString(metadata.ReceiverPublicKey),
	}
}