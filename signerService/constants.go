package main

import (
	"os"
	"time"

	flow_aws_kms "bitbucket.org/carsonliving/aws-kms-client"
	ep "bitbucket.org/carsonliving/cryptographymodules/ecdsaoperations"
	ed "bitbucket.org/carsonliving/cryptographymodules/eddsaoperations"
	log "github.com/sirupsen/logrus"
)

var (
	Database           string
	Collection         string
	SignerService      ep.ECDSAWalletSigning
	WsSignerService    ep.WSECDSAWalletSigning
	EDDSASignerService ed.EDDSAWallet
	ECDSAWalletService ep.ECDSAWallet
)

// Retry wait time
const RetrySleep = 3 * time.Second

// AWS KSM Client instance
var KSM flow_aws_kms.KMSClient = KSMClient()

// TODO: need configuration file with list of ERC20 tokens we want to support
// need symbol and smart contract address (smart contract address will be chain specific main vs test)
var Erc20Map map[string]string = map[string]string{
	"QKC": "0xb2a28A6f755b85eeF3cD41058A5d2A7A398281FC",
}

// Array of support blockchains
var BlockchainIds = []string{
	"ETH",
	"BTC",
	"ADA",
	"ALGO",
	"AVAX",
}

const (
	//pubkeyCompressed   byte = 0x2 // y_bit + x coord
	PubkeyUncompressed byte = 0x4 // x coord + y coord
	//pubkeyHybrid       byte = 0x6 // y_bit + x coord + y coord
)

// Recovery Flow
const (
	Initiate         = "initiate"
	Initiated        = "initiated"
	MFAVerified      = "mfaVerified"
	CustomerVerified = "customerVerified"
	Complete         = "complete"
	Recovery         = "RecoveryRecord"
)

var MongoDatabase string

func getDatabase() {
	var ok bool
	MongoDatabase, ok = os.LookupEnv("MONGODB_DATABASE")
	if !ok {
		log.Fatal("missing environment variable: MONGODB_DATABASE")
	}
	return
}
