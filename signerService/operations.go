package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	//"github.com/coinbase/kryptology/pkg/core/curves"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostECDSAKeyShare api function for receiving and storing participant
// keyshare information
func PostECDSAKeyShare(c *gin.Context) {
	var keyShare KeyShare
	err := json.NewDecoder(c.Request.Body).Decode(&keyShare)
	if err != nil {
		log.Error("Errow decoding share data: ", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
		return
	}
	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")
	defer CloseClientDB(DB)

	//check if key share already exists
	_, err = readECDSAShare(keyShare.UserId, keyShare.BlockchainId, keyShare.AccountName, keyShareCollection)
	if err == mongo.ErrNoDocuments {
		err = writeECDSAShare(keyShare, keyShareCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}

		err = createAccountRecord(keyShare.UserId, keyShare.BlockchainId, keyShare.AccountName, keyShare.Address, userCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
	} else if err != nil {
		WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
	} else {
		err = updateECDSAShare(keyShare, keyShareCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
	}

	ValidateAndWriteResponse("Success", err, c.Writer)
	return
}

// PostECDSAKeyShare api function for receiving and storing participant
// keyshare information
func PostEDDSAKeyShare(c *gin.Context) {
	var keyShare EDDSAShare
	err := json.NewDecoder(c.Request.Body).Decode(&keyShare)
	if err != nil {
		log.Error("Error decoding share data: ", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
		return
	}
	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")
	defer CloseClientDB(DB)

	_, err = readEDDSAShare(keyShare.UserId, keyShare.BlockchainId, keyShare.AccountName, keyShareCollection)
	if err == mongo.ErrNoDocuments {
		err = writeEDDSAShare(keyShare, keyShareCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}

		err = createAccountRecord(keyShare.UserId, keyShare.BlockchainId, keyShare.AccountName, keyShare.Address, userCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
	} else if err != nil {
		WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
	} else {
		err = updateEDDSAShare(keyShare, keyShareCollection)
		if err != nil {
			WriteErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
	}

	ValidateAndWriteResponse("Success", err, c.Writer)
	return
}

// GetECDSAKeyShare returns the key share after permissioning protocol
func GetECDSAKeyShare(c *gin.Context) {
	userId := c.Param("userId")
	blockchainId := c.Param("blockchainId")
	accountName := c.Param("accountName")
	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")
	defer CloseClientDB(DB)

	recoveryRecord, err := readRecoveryRecord(userId, userCollection)
	if err != nil {
		log.Error("Error getting recovery record err:", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Recovery not initiated"), c.Writer)
		return
	}
	if recoveryRecord.Status == "customerVerified" {
		share, err := readECDSAShare(userId, blockchainId, accountName, keyShareCollection)
		if err != nil {
			log.Error("Error reading share err:", err)
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}

		ValidateAndWriteResponse(share, err, c.Writer)
		return
	}

	log.Error("Recovery not verified")
	WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Recovery not verified"), c.Writer)
	return
}

// GetECDSAKeyShare returns the key share after permissioning protocol
func GetEDDSAKeyShare(c *gin.Context) {
	userId := c.Param("userId")
	blockchainId := c.Param("blockchainId")
	accountName := c.Param("accountName")
	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")
	defer CloseClientDB(DB)

	recoveryRecord, err := readRecoveryRecord(userId, userCollection)
	if err != nil {
		log.Error("Error getting recovery record err:", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Recovery not initiated"), c.Writer)
		return
	}
	if recoveryRecord.Status == "customerVerified" {
		share, err := readEDDSAShare(userId, blockchainId, accountName, keyShareCollection)
		if err != nil {
			log.Error("Error reading share err:", err)
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}

		ValidateAndWriteResponse(share, err, c.Writer)
		return
	}

	log.Error("Recovery not verified")
	WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Recovery not verified"), c.Writer)
	return
}

// GetUserAccounts returns the accounts for a user
func GetUserAccounts(c *gin.Context) {
	userId := c.Param("userId")
	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	defer CloseClientDB(DB)
	var result []AccountRecord

	for _, blockchainId := range BlockchainIds {
		accountRecors, err := readAccountRecords(userId, blockchainId, userCollection)
		if err != nil {
			log.Error("Error reading account record err:", err)
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
		result = append(result, accountRecors...)
	}

	ValidateAndWriteResponse(result, nil, c.Writer)
	return
}

func GetRecoverUserAccountsStatus(c *gin.Context) {
	userId := c.Param("userId")

	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
	defer CloseClientDB(DB)
	recoveryRecord, err := readRecoveryRecord(userId, userCollection)

	if err == mongo.ErrNoDocuments {
		ValidateAndWriteResponse(SuccessDetails{
			Message: "not initiated",
		}, nil, c.Writer)
		return
	}

	if err != nil {
		log.Error("Error getting recovery record err:", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
		return
	}

	ValidateAndWriteResponse(SuccessDetails{
		Message: recoveryRecord.Status,
	}, err, c.Writer)
	return
}

// RecoverUserAccounts creates or updates recovery record
func RecoverUserAccounts(c *gin.Context) {
	userId := c.Param("userId")

	state := c.Query("state")
	log.Info("State passed:", state)

	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")

	defer CloseClientDB(DB)

	if state == "" {
		log.Error("Missing state query parameter")
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Missing state query parameter"), c.Writer)

		return
	} else if state == Initiate {
		err := createRecoveryRecord(userId, userCollection)
		if err != nil {
			log.Error("Error creating recovery record err:", err)
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
	} else {
		recoveryRecord, err := readRecoveryRecord(userId, userCollection)
		if err != nil {
			log.Error("Error getting recovery record err:", err)
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
			return
		}
		if state == CustomerVerified && recoveryRecord.Status == Initiated {
			err := updateRecoveryRecord(userId, state, userCollection)
			if err != nil {
				log.Error("Error creating recovery record err:", err)
				WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
				return
			}
		} else if state == Complete && recoveryRecord.Status == CustomerVerified {
			err := deleteRecoveryRecord(userId, userCollection)
			if err != nil {
				log.Error("Error deleting recovery record err:", err)
				WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
				return
			}
		} else {
			log.Error("Error updating recovery record: wrong status")
			WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", "Error updating recovery record: wrong status"), c.Writer)
			return
		}

	}

	ValidateAndWriteResponse(SuccessDetails{
		Message: "Success",
	}, nil, c.Writer)
	return
}

// POSTEDDSASignature completes the signing flow for eddsa keys
func POSTEDDSASignature(c *gin.Context) {

	userId := c.Param("userId")
	blockchainId := c.Param("blockchainId")
	accountName := c.Param("accountName")
	//messageHash := r.URL.Query().Get("messageHash")

	var txInput TxInputs
	err := json.NewDecoder(c.Request.Body).Decode(&txInput)
	if err != nil {
		log.Error("Error decoding share data: ", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
		return
	}

	var DB *mongo.Client = ConnectDB()
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")

	share, err := readEDDSAShare(userId, blockchainId, accountName, keyShareCollection)
	if err != nil {
		log.Error("Error reading share err:", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
	}

	fullSignature, err := EDDSASignerService.CustodySign(share.PK, share.SigShare, txInput.Hash, txInput.Signature)
	if err != nil {
		log.Error("Error generating full signature:", err)
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error: %s", err), c.Writer)
	}

	ValidateAndWriteResponse(fullSignature, err, c.Writer)
	return
}

func generatePaillierKeys() {

	for {

		id := uuid.New().String()
		var DB *mongo.Client = ConnectDB()
		paillierKeyCollection := DB.Database(MongoDatabase).Collection("PaillierKeyCollection")
		paillierKey, err := ECDSAWalletService.CreatePaillierKeyPair()
		if err != nil {
			log.Error("Error creating paillier key", err)
			return
		}

		keyData := PaillierKey{
			UniqueId:        id,
			PaillierKeyData: paillierKey,
		}
		errDB := writePaillierKey(keyData, paillierKeyCollection)
		if errDB != nil {
			log.Error("Error writing paillier key error:", errDB)
			return
		}
		defer CloseClientDB(DB)

		time.Sleep(60 * time.Second)
	}
}

func RequestPaillierKey(c *gin.Context) {
	userId := c.Param("userId")
	id := uuid.New().String()

	for i := 1; i < 3; i++ {
		go func(i int, id string, userId string) {
			var DB *mongo.Client = ConnectDB()
			userCollection := DB.Database(MongoDatabase).Collection("UserCollection")
			paillierKey, err := ECDSAWalletService.CreatePaillierKeyPair()
			if err != nil {
				log.Error("Error creating paillier key", err)
				keyData := PaillierKey{
					UserId:    userId,
					UniqueId:  id,
					KeyNumber: i,
					Error:     err.Error(),
				}
				errDB := writePaillierKey(keyData, userCollection)
				log.Error("Error writing paillier key error:", errDB)
				return
			}

			keyData := PaillierKey{
				UserId:          userId,
				UniqueId:        id,
				KeyNumber:       i,
				PaillierKeyData: paillierKey,
			}
			errDB := writePaillierKey(keyData, userCollection)
			log.Error("Error writing paillier key error:", errDB)
			defer CloseClientDB(DB)
		}(i, id, userId)
	}

	ValidateAndWriteResponse(SuccessDetails{
		Message: id,
	}, nil, c.Writer)

}

func GetPaillierKey(c *gin.Context) {
	userId := c.Param("userId")
	id := c.Query("id")
	if id == "" {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Missing query parameter: id"), c.Writer)
		return
	}

	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")

	defer CloseClientDB(DB)

	paillierKeys, err := readPaillierKeys(userId, id, userCollection)

	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error getting paillier keys: %s", err), c.Writer)
		return
	}
	if len(paillierKeys) == 0 {
		ValidateAndWriteResponse(SuccessDetails{
			Message: "No paillier keys found",
		}, err, c.Writer)
		return
	} else if len(paillierKeys) == 1 {
		ValidateAndWriteResponse(SuccessDetails{
			Message: "Only one paillier key found",
		}, err, c.Writer)
		return
	}

	paillierKeysJSON, err := json.Marshal(paillierKeys)
	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error json marshall: %s", err), c.Writer)
		return
	}

	ValidateAndWriteResponse(SuccessDetails{
		Message: string(paillierKeysJSON),
	}, err, c.Writer)

	return
}

func RemovePaillierKey(c *gin.Context) {
	userId := c.Param("userId")
	id := c.Query("id")
	if id == "" {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Missing query parameter: id"), c.Writer)
		return
	}

	var DB *mongo.Client = ConnectDB()
	userCollection := DB.Database(MongoDatabase).Collection("UserCollection")

	defer CloseClientDB(DB)

	err := deletePaillierKeys(userId, id, userCollection)

	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error getting paillier keys: %s", err), c.Writer)
		return
	}

	ValidateAndWriteResponse(SuccessDetails{
		Message: "Success",
	}, err, c.Writer)

	return
}

func GetRandomPaillierKey(c *gin.Context) {
	var DB *mongo.Client = ConnectDB()
	paillierKeyCollection := DB.Database(MongoDatabase).Collection("PaillierKeyCollection")

	defer CloseClientDB(DB)

	paillierKeys, err := readRandomPaillierKeys(paillierKeyCollection)

	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error getting paillier keys: %s", err), c.Writer)
		return
	}
	if len(paillierKeys) == 0 {
		ValidateAndWriteResponse(SuccessDetails{
			Message: "No paillier keys found",
		}, err, c.Writer)
		return
	} else if len(paillierKeys) == 1 {
		ValidateAndWriteResponse(SuccessDetails{
			Message: "Only one paillier key found",
		}, err, c.Writer)
		return
	}

	paillierKeysJSON, err := json.Marshal(paillierKeys)
	if err != nil {
		WriteErrorResponse(http.StatusBadRequest, fmt.Sprintf("Error json marshall: %s", err), c.Writer)
		return
	}

	log.Info("Paillier keys: ", string(paillierKeysJSON))

	ValidateAndWriteResponse(SuccessDetails{
		Message: string(paillierKeysJSON),
	}, err, c.Writer)

	return
}
