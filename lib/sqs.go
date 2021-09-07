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

type SQSQueueTransaction struct {
	// A string that uniquely identifies this transaction. This is a sha256 hash
	// of the transaction’s data encoded using base58 check encoding.
	TransactionIDBase58Check string
	// The raw hex of the transaction data. This can be fully-constructed from
	// the human-readable portions of this object.
	RawTransactionHex string `json:",omitempty"`
	// The inputs and outputs for this transaction.
	// The signature of the transaction in hex format.
	SignatureHex string `json:",omitempty"`
	// Will always be “0” for basic transfers
	TransactionType string `json:",omitempty"`
	// TODO: Create a TransactionMeta portion for the response.

	// The hash of the block in which this transaction was mined. If the
	// transaction is unconfirmed, this field will be empty. To look up
	// how many confirmations a transaction has, simply plug this value
	// into the "block" endpoint.
	BlockHashHex string `json:",omitempty"`

	TransactionMetadata TransactionMetadata `json:",omitempty"`
}

func NewSQSQueue(client *sqs.Client, queueUrl string) *SQSQueue {
	newSqsQueue := SQSQueue{}
	newSqsQueue.sqsClient = client
	newSqsQueue.queueUrl = &queueUrl
	return &newSqsQueue
}

func (sqsQueue *SQSQueue) SendTxnMessage(txn *MsgBitCloutTxn, txnMeta *TransactionMetadata, params *BitCloutParams) {
	response := APITransactionToResponse(*txn, *txnMeta, *params)
	sendMessageInput := &sqs.SendMessageInput{
		DelaySeconds: 10,
		MessageBody: aws.String(string(serialize(response))),
		QueueUrl:    sqsQueue.queueUrl,
	}
	_, err := sqsQueue.sqsClient.SendMessage(context.TODO(), sendMessageInput);
	if err != nil {
		glog.Error(err)
	}
}

func serialize(sqsQueueTxn *SQSQueueTransaction) []byte {
	data := make(map[string]string)
	data["transactionType"] = sqsQueueTxn.TransactionType
	data["transactionHex"] = sqsQueueTxn.RawTransactionHex
	data["transactionAffectedPublicKey0"] = sqsQueueTxn.TransactionMetadata.AffectedPublicKeys[0].PublicKeyBase58Check
	data["transactionAffectedPublicKey0Metadata"] = sqsQueueTxn.TransactionMetadata.AffectedPublicKeys[0].Metadata

	jsonData, err := json.Marshal(data)
	if err != nil {
		glog.Error("Could not serialize SQS Queue input")
	}
	return jsonData

}

func APITransactionToResponse(
	txnn MsgBitCloutTxn,
	txnMeta TransactionMetadata,
	params BitCloutParams) *SQSQueueTransaction {

	signatureHex := ""
	if txnn.Signature != nil {
		signatureHex = hex.EncodeToString(txnn.Signature.Serialize())
	}

	txnBytes, _ := txnn.ToBytes(false /*preSignature*/)
	ret := &SQSQueueTransaction{
		TransactionIDBase58Check: PkToString(txnn.Hash()[:], &params),
		RawTransactionHex:        hex.EncodeToString(txnBytes),
		SignatureHex:             signatureHex,
		TransactionType:          txnn.TxnMeta.GetTxnType().String(),

		TransactionMetadata: txnMeta,
	}

	return ret
}