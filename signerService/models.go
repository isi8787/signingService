package main

import (
	"encoding/json"

	ep "bitbucket.org/carsonliving/cryptographymodules/ecdsaoperations"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invoke Request is for picking up messages from azure bus
type InvokeRequest struct {
	Data     map[string]json.RawMessage //content of message
	Metadata map[string]interface{}     //metadata associated with message
}

// KeyShare represents stored key share data and identifying information
type KeyShare struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"` //mongoDB object id created when item inserted to DB
	UserId       string              `bson:"userId"`        // userId created during registration in active directory
	TokenId      string              `bson:"tokenId"`       // tokenId is the symbol for specific asset ETH, BTC, ERC20 symbol
	AccountName  string              `bson:"accountName"`   // accountName is the user defined nickname for a specific set of credentials
	BlockchainId string              `bson:"blockchainId"`  // blockchainId is the symbol for a specific blockchain, this is used to link credentials from L1 to ERC20 or ERC721 tokens
	Address      string              `bson:"address,omitempty"`
	ShareData    ep.ECDSAParticipant `bson:"shareData"` // shareData is participant specific MPC sensitive data
}

// MessageToSigner struct for message sent from aggregator to signer
type MessageToSigner struct {
	Message   string `json:"message,omitempty"`   // message is payload transmitted over azure bus
	Operation string `json:"operation,omitempty"` // operation is identifier of how message should be processed
	ObjectID  string `json:"objectID,omitempty"`  // objectId is unique identifier that message relates to
	UserId    string `json:"userId"`              // userId associated with action and objectid
}

// MessageToAggregator struct for sending messages from signer to aggegator
type MessageToAggregator struct {
	Signer    string `json:"signer,omitempty"`    // signer is the message originator
	Message   string `json:"message,omitempty"`   // message is payload transmitted over azure bus
	Operation string `json:"operation,omitempty"` // operation is identifier of how message should be processed
	ObjectID  string `json:"objectID,omitempty"`  // objectId is unique identifier that message relates to
	UserId    string `json:"userId"`              // userId associated with action and objectid
}

// TXState represents a tx state during rounds
type TXState struct {
	MessageHash string `bson:"messageHash"` // messageHash is unique identifier for processing a transaction
	State       string `bson:"state"`       // state is the state for the MPC signing process
	Status      string `bson:"status"`      // status of a transaction being processed by signing service
	UserId      string `bson:"userId"`      // userId associated with action and objectid
}

// BroadcastMessage is standard struct for sending aggregate round broadcast
type BroadcastMessage struct {
	MessageHash string `json:"messageHash"`          // messageHash is unique identifier for processing a transaction
	UserId      string `json:"userId"`               // userId associated with action and objectid
	TokenId     string `json:"tokenId"`              // tokenId is the symbol for specific asset ETH, BTC, ERC20 symbol
	Broadcast   string `json:"broadcast, omitempty"` //broadcast is the message being transmitted to multiple signers at once
}

// BasicTx is structure for sending tokens from one address to another
type BasicTx struct {
	UserId       string `bson:"userId"`       // userId associated with account
	BlockchainId string `bson:"blockchainId"` // blockchainId is the symbol for a specific blockchain, this is used to link credentials from L1 to ERC20 or ERC721 tokens
	TokenId      string `bson:"tokenId"`      // tokenId is the symbol for specific asset ETH, BTC, ERC20 symbol
	AccountName  string `bson:"accountName"`  // accountName is the user defined nickname for a specific set of credentials
	Value        string `bson:"value"`        // value being transfered by transaction to target account
	ToAddress    string `bson:"toAddress"`    // toAddress is recipient hex encoded address
	FullTx       string `bson:"fullTx"`       // fullTx is the JSON stringified transaction that is submitted to target blockchain
	TxHash       string `bson:"txHash"`       // txHash is the tx hash used to identify transaction
	Status       string `bson:"status"`       // status tracks transaction status from generation to signing to completion
}

// ETHAccounts is a structure for returning to wallet the balances associated with a users ETH blockchain holdings
type ETHAccounts struct {
	Address       string            `json:"address"`       // hex string address on ETH
	Balance       string            `json:"balance"`       // ETH balance
	ERC20Balances map[string]string `json:"erc20Balances"` // ERC20 balance for supported tokens
}

type SigningRounds struct {
	Round      string // tracks which ECDSA MPC round is being processed
	Identifier string // 1 is for custody service, 2 is for escrow service, 3 is for mobile or other client
	Message    string // signing broadcast from other participant during ECDSA MPC rounds
}

type AccountRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"` //mongoDB object id created when item inserted to DB
	UserId       string             `bson:"userId"`        // userId created during registration in active directory
	AccountName  string             `bson:"accountName"`   // accountName is the user defined nickname for a specific set of credentials
	BlockchainId string             `bson:"blockchainId"`  // blockchainId is the symbol for a specific blockchain, this is used to link credentials from L1 to ERC20 or ERC721 tokens
	Address      string             `bson:"address,omitempty"`
	RecordType   string             `bson:"recordType"` // extra field to improve searching
}

type EDDSAShare struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`          //mongoDB object id created when item inserted to DB
	UserId       string             `bson:"userId,omitempty"`       // userId created during registration in active directory
	AccountName  string             `bson:"accountName,omitempty"`  // accountName is the user defined nickname for a specific set of credentials
	BlockchainId string             `bson:"blockchainId,omitempty"` // blockchainId is the symbol for a specific blockchain, this is used to link credentials from L1 to ERC20 or ERC721 tokens
	TokenId      string             `bson:"tokenId,omitempty"`      // tokenId
	Address      string             `bson:"address,omitempty"`
	PK           string             `bson:"pk,omitempty"`
	SigShare     string             `bson:"sigShare,omitempty"`
	EscrowShare  string             `bson:"escrowShare,omitempty"`
}

type TxInputs struct {
	Hash       string `json:"hash"`
	InputIndex int    `json:"inputIndex"`
	Signature  string `json:"signature"`
}

type RecoveryRecord struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`  //mongoDB object id created when item inserted to DB
	UserId         string             `bson:"userId"`         // userId created during registration in active directory
	AccountRecords []AccountRecord    `bson:"accountRecords"` // accountrecord that needs to be processed
	Status         string             `bson:"status"`         // status of account record being recovered
	RecordType     string             `bson:"recordType"`     // extra field to improve searching
}

type SuccessDetails struct {
	Message string `json:"message"`
}

type PaillierKey struct {
	UniqueId        string `bson:"uniqueId"`
	UserId          string `bson:"userId"`
	PaillierKeyData string `bson:"paillierKeyData"`
	KeyNumber       int    `bson:"keyNumber"`
	Error           string `bson:"error"`
}

type ApiError struct {
	Status bool         `json:"status"`
	Err    ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ApiSuccess struct {
	Status bool        `json:"status"`
	Result interface{} `json:"result"`
}

type AWSStorage struct {
	Index string   `bson:"index"`
	Value []string `bson:"value"`
}
