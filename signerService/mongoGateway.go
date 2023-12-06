package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"

	flow_aws_kms "bitbucket.org/carsonliving/aws-kms-client"
)

// AWS KSM Client instance
func KSMClient() flow_aws_kms.KMSClient {
	RegionDeploy, ok := os.LookupEnv("RegionDeploy")
	if !ok {
		log.Fatal("missing environment variable: CustodyServiceKSMKey	")
	}

	KSMKey, ok := os.LookupEnv("CustodyServiceKSMKey")
	if !ok {
		log.Fatal("missing environment variable: CustodyServiceKSMKey	")
	}
	return flow_aws_kms.GetWithDefaultConfig(RegionDeploy, KSMKey)
}

// ConnectDB creates a MongoDB client
func ConnectDB() *mongo.Client {
	mongoDBConnectionString, ok := os.LookupEnv("CosmosDbConnectionString")
	if !ok {
		log.Fatal("missing environment variable: CosmosDbConnectionString")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoDBConnectionString).SetDirect(true)
	c, err := mongo.NewClient(clientOptions)

	err = c.Connect(ctx)

	if err != nil {
		log.Error("unable to initialize connection ", err)
	}
	err = c.Ping(ctx, nil)
	if err != nil {
		log.Error("unable to connect ", err)
	}

	return c
}

// Disconnect client
func CloseClientDB(client *mongo.Client) {
	if client == nil {
		return
	}

	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Connection to MongoDB closed.")
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func writeECDSAShare(dataEntry KeyShare, keyShareCollection *mongo.Collection) error {
	keyvaultindex := dataEntry.UserId + "-" + dataEntry.BlockchainId + "-" + dataEntry.AccountName // TODO: need to salt hash to create a more obfuscated index
	dataEntryJSON, err := json.Marshal(dataEntry)
	if err != nil {
		log.Error("Error encoding json:", err)
	}

	ciphertexts, err := chunkEncryptData(dataEntryJSON)
	if err != nil {
		log.Error("Error encrypting data:", err)
	}

	awsKeyObject := AWSStorage{
		Index: keyvaultindex,
		Value: ciphertexts,
	}

	ctx := context.Background()
	keyShare, err := keyShareCollection.InsertOne(ctx, awsKeyObject)
	if err != nil {
		log.Error("failed to add to db:", err)
		return err
	}

	log.Info("Created KeyShare:", keyShare)

	return nil
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func updateECDSAShare(dataEntry KeyShare, keyShareCollection *mongo.Collection) error {
	keyvaultindex := dataEntry.UserId + "-" + dataEntry.BlockchainId + "-" + dataEntry.AccountName // TODO: need to salt hash to create a more obfuscated index
	dataEntryJSON, err := json.Marshal(dataEntry)
	if err != nil {
		log.Error("Error encoding json:", err)
	}

	ciphertexts, err := chunkEncryptData(dataEntryJSON)
	if err != nil {
		log.Error("Error encrypting data:", err)
	}

	awsKeyObject := AWSStorage{
		Index: keyvaultindex,
		Value: ciphertexts,
	}

	ctx := context.Background()
	filter := bson.M{"index": keyvaultindex}
	update := bson.M{"$set": awsKeyObject}
	_, err = keyShareCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error("failed to update to db:", err)
		return err
	}

	log.Info("Update KeyShare:", dataEntry)

	return nil
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func writeEDDSAShare(dataEntry EDDSAShare, keyShareCollection *mongo.Collection) error {
	keyvaultindex := dataEntry.UserId + "-" + dataEntry.BlockchainId + "-" + dataEntry.AccountName // TODO: need to salt hash to create a more obfuscated index
	dataEntryJSON, _ := json.Marshal(dataEntry)
	ciphertexts, err := chunkEncryptData(dataEntryJSON)
	if err != nil {
		log.Error("Error encrypting data:", err)
	}

	awsKeyObject := AWSStorage{
		Index: keyvaultindex,
		Value: ciphertexts,
	}

	ctx := context.Background()
	keyShare, err := keyShareCollection.InsertOne(ctx, awsKeyObject)
	if err != nil {
		log.Error("failed to add to db:", err)
		return err
	}

	log.Info("Created KeyShare:", keyShare)

	return nil
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func updateEDDSAShare(dataEntry EDDSAShare, keyShareCollection *mongo.Collection) error {
	keyvaultindex := dataEntry.UserId + "-" + dataEntry.BlockchainId + "-" + dataEntry.AccountName // TODO: need to salt hash to create a more obfuscated index
	dataEntryJSON, err := json.Marshal(dataEntry)
	if err != nil {
		log.Error("Error encoding json:", err)
	}

	ciphertexts, err := chunkEncryptData(dataEntryJSON)
	if err != nil {
		log.Error("Error encrypting data:", err)
	}

	awsKeyObject := AWSStorage{
		Index: keyvaultindex,
		Value: ciphertexts,
	}

	ctx := context.Background()
	filter := bson.M{"index": keyvaultindex}
	update := bson.M{"$set": awsKeyObject}
	_, err = keyShareCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error("failed to update to db:", err)
		return err
	}

	log.Info("Update KeyShare:", dataEntry)

	return nil
}

// chunk input data into 4kb chunks
func chunkEncryptData(dataEntryJSON []byte) ([]string, error) {
	// Split the plaintext data into chunks of 4 KB or less
	chunkSize := 4096
	var ciphertexts []string
	for i := 0; i < len(dataEntryJSON); i += chunkSize {
		end := i + chunkSize
		if end > len(dataEntryJSON) {
			end = len(dataEntryJSON)
		}
		chunk := dataEntryJSON[i:end]

		// Call the Encrypt operation
		encryptedChunk, err := KSM.Encrypt(chunk)
		if err != nil {
			return ciphertexts, err
		}

		encryptedbse64 := base64.StdEncoding.EncodeToString(encryptedChunk)

		// Append the ciphertext to the slice of ciphertexts
		ciphertexts = append(ciphertexts, encryptedbse64)
	}
	return ciphertexts, nil
}

// decrypChunkData
func decrypChunkData(ciphertexts []string) ([]byte, error) {
	var plaintexts [][]byte
	for i := 0; i < len(ciphertexts); i++ {
		decodeb64, err := base64.StdEncoding.DecodeString(ciphertexts[i])
		if err != nil {
			return nil, err
		}

		// Call the Decrypt operation
		output, err := KSM.Decrypt(decodeb64)
		if err != nil {
			return nil, err
		}

		// Append the plaintext to the slice of plaintexts
		plaintexts = append(plaintexts, output)
	}

	plaintext := bytes.Join(plaintexts, []byte(""))
	return plaintext, nil
}

// createAccountRecord saves record of index used to store keyShare
func createAccountRecord(userId, blockchainId, accountName, address string, todoCollection *mongo.Collection) error {
	_, err := readAccount(userId, blockchainId, accountName, todoCollection)
	if err == mongo.ErrNoDocuments {
		ctx := context.Background()
		account, err := todoCollection.InsertOne(ctx, AccountRecord{UserId: userId, BlockchainId: blockchainId, AccountName: accountName, Address: address, RecordType: "AccountData"})
		if err != nil {
			log.Error("failed to add to db:", err)
			return err
		}
		log.Info("Created Account:", account)
	}

	return nil
}

// readAccount retrieve account record
func readAccount(userId, blockchainId, accountName string, todoCollection *mongo.Collection) (AccountRecord, error) {
	var res AccountRecord
	filter := bson.M{"recordType": "AccountData", "userId": userId, "blockchainId": blockchainId, "accountName": accountName}

	ctx := context.Background()
	err := todoCollection.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		log.Error("Error reading record from db err:", err)
		return res, err
	}

	log.Info("The account data is: ", res)

	return res, nil
}

// readAccountRecords retrieve records of indexes used to store keyShare
func readAccountRecords(userId, blockchainId string, todoCollection *mongo.Collection) ([]AccountRecord, error) {
	var res []AccountRecord
	filter := bson.M{"recordType": "AccountData", "userId": userId, "blockchainId": blockchainId}

	ctx := context.Background()
	listRes, err := todoCollection.Find(ctx, filter)
	if err != nil {
		log.Error("Error reading  record from db err:", err)
		return res, fmt.Errorf("Error reading record from db err: %s", err)
	}

	if err = listRes.All(ctx, &res); err != nil {
		log.Fatal(err)
	}

	log.Info("The account data is: ", res)
	defer listRes.Close(ctx)

	return res, nil
}

// deleteAccount deletes an account entry
func deleteAccount(userId, blockchainId, accountName string, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.D{{"recordType", "AccountData"}, {"userId", userId}, {"blockchainId", blockchainId}, {"accountName", accountName}}
	_, err := todoCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("failed to delete account ", err)
		return err
	}
	return nil
}

// readShare return target key share based on userid , blockchainid and accountName
func readECDSAShare(userId string, blockchainId string, accountName string, keyShareCollection *mongo.Collection) (KeyShare, error) {
	var keyShare KeyShare
	keyvaultindex := userId + "-" + blockchainId + "-" + accountName // TODO: need to salt hash to create a more obfuscated index

	var awsKeyObject AWSStorage
	filter := bson.M{"index": keyvaultindex}

	ctx := context.Background()
	err := keyShareCollection.FindOne(ctx, filter).Decode(&awsKeyObject)
	if err != nil {
		log.Error("Error reading record from db err:", err)
		return keyShare, err
	}

	keySharebytes, err := decrypChunkData(awsKeyObject.Value)
	if err != nil {
		log.Error("Error decrypting key share from key vault ", err)
		return keyShare, err
	}

	err = json.Unmarshal(keySharebytes, &keyShare)
	if err != nil {
		log.Error("Error decoding key share from key vault ", err)
	}

	return keyShare, nil
}

// readShare return target key share based on userid , blockchainid and accountName
func readEDDSAShare(userId string, blockchainId string, accountName string, keyShareCollection *mongo.Collection) (EDDSAShare, error) {
	var keyShare EDDSAShare
	keyvaultindex := userId + "-" + blockchainId + "-" + accountName // TODO: need to salt hash to create a more obfuscated index

	var awsKeyObject AWSStorage
	filter := bson.M{"index": keyvaultindex}

	ctx := context.Background()
	err := keyShareCollection.FindOne(ctx, filter).Decode(&awsKeyObject)
	if err != nil {
		log.Error("Error reading record from db err:", err)
		return keyShare, err
	}

	keySharebytes, err := decrypChunkData(awsKeyObject.Value)
	if err != nil {
		log.Error("Error decrypting key share from key vault ", err)
		return keyShare, err
	}

	err = json.Unmarshal(keySharebytes, &keyShare)
	if err != nil {
		log.Error("Error decoding key share from key vault ", err)
	}

	return keyShare, nil
}

// writeState saves tx state during ecdsa rounds
func writeState(state, messageHash, status, userId string, todoCollection *mongo.Collection) (interface{}, error) {
	ctx := context.Background()
	_, err := todoCollection.InsertOne(ctx, TXState{State: state, Status: status, MessageHash: messageHash, UserId: userId})
	if err != nil {
		log.Error("failed to add todo ", err)
		return nil, err
	}
	return nil, nil
}

// readState reads saved tx state during ecdsa rounds
func readState(messageHash, userId string, todoCollection *mongo.Collection) (TXState, error) {
	var res TXState
	var filter interface{}
	filter = bson.D{{"messageHash", messageHash}}

	ctx := context.Background()
	err := todoCollection.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		log.WithFields(log.Fields{"messageHash": messageHash}).Error("Error reading messageHash from db err: ", err)
		return res, fmt.Errorf("Error reading  msg:%s from db err: %s", messageHash, err)
	}
	return res, nil
}

// updateState saves tx state
func updateState(state, messageHash, status, userId string, todoCollection *mongo.Collection) (interface{}, error) {
	ctx := context.Background()
	filter := bson.D{{"messageHash", messageHash}, {"userId", userId}}
	update := bson.M{
		"$set": TXState{State: state, Status: status, MessageHash: messageHash, UserId: userId},
	}
	_, err := todoCollection.UpdateOne(ctx, filter, update)
	if err != nil {

		log.Error("failed to update todo ", err)
		return nil, err
	} else {
		log.Info("updated DB for :", messageHash)
	}

	return nil, nil
}

// retryDB retry will re-run the given function if failed till attempts. Between each attempt, sleep a while
func retryDB(attempts int, sleep time.Duration, state, messageHash, status, userId string, todoCollection *mongo.Collection, fn func(string, string, string, string, *mongo.Collection) (interface{}, error)) (result interface{}, err error) {
	for i := 0; i < attempts; i++ {
		result, err := fn(state, messageHash, status, userId, todoCollection)
		if err != nil {
			log.Error("Retrying after error: ", err)
			time.Sleep(sleep)
			continue
		}
		log.WithFields(log.Fields{"result": result}).Info("Got result, will exit the retry")
		return result, nil
	}
	return nil, fmt.Errorf("retry failed after %d attempts, last error: %s", attempts, err)
}

// deleteState deletes an entry for an MPC state associated with message hash and user id
func deleteState(messageHash, userId string, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.D{{"messageHash", messageHash}, {"userId", userId}}
	_, err := todoCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("failed to delete state ", err)
		return err
	}
	return nil
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func writeTx(dataEntry BasicTx, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	_, err := todoCollection.InsertOne(ctx, dataEntry)
	if err != nil {
		log.Error("failed to add BasicTx ", err)
		return err
	}
	return nil
}

// readTx return target key share based on userid and token id
func readTx(txHash string, todoCollection *mongo.Collection) (BasicTx, error) {
	var res BasicTx
	var filter interface{}
	filter = bson.D{{"txHash", txHash}}

	ctx := context.Background()
	err := todoCollection.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to find tx with messageHAsh:%s, err: %v", txHash, err)
	}

	return res, nil
}

// updateState saves tx state
func updateTx(txHash string, tx BasicTx, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.D{{"txHash", txHash}}
	update := bson.M{
		"$set": tx,
	}
	_, err := todoCollection.UpdateOne(ctx, filter, update)
	if err != nil {

		log.Error("failed to update todo ", err)
		return err
	} else {
		log.Info("updated DB for: ", txHash)
	}

	return nil
}

// createRecoveryRecord saves record of users recovery flow
func createRecoveryRecord(userId string, todoCollection *mongo.Collection) error {
	_, noDocs := readRecoveryRecord(userId, todoCollection)
	if noDocs == mongo.ErrNoDocuments {
		var accountRecords []AccountRecord
		for _, blockchainId := range BlockchainIds {
			accountRecord, err := readAccountRecords(userId, blockchainId, todoCollection)
			if err != nil {
				log.Error("Error reading account record err:", err)
				return fmt.Errorf("Error checking for previous recovery record: %s", err)
			}
			accountRecords = append(accountRecords, accountRecord...)
		}

		ctx := context.Background()

		_, err := todoCollection.InsertOne(ctx, RecoveryRecord{UserId: userId, AccountRecords: accountRecords, Status: Initiated, RecordType: Recovery})
		if err != nil {
			log.Error("Failed to add recovery record to db:", err)
			return err
		}
		return nil
	}

	return fmt.Errorf("Recovery process already started")
}

// updateRecoveryRecord saves updated recovery record
func updateRecoveryRecord(userId, status string, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.M{"recordType": Recovery, "userId": userId}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := todoCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error("failed to update todo ", err)
		return err
	}

	return nil
}

// readRecoveryRecord retrieve record of users recovery flow
func readRecoveryRecord(userId string, todoCollection *mongo.Collection) (RecoveryRecord, error) {
	var res RecoveryRecord
	filter := bson.M{"recordType": Recovery, "userId": userId}

	ctx := context.Background()
	err := todoCollection.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		log.Error("Error reading record from db err:", err)
		return res, err
	}

	log.Info("The account data is: ", res)

	return res, nil
}

// readRecoveryRecords retrieve all records of recovery
func readRecoveryRecords(userId, blockchainId string, todoCollection *mongo.Collection) ([]RecoveryRecord, error) {
	var res []RecoveryRecord
	filter := bson.M{"recordType": Recovery, "userId": userId}

	ctx := context.Background()
	listRes, err := todoCollection.Find(ctx, filter)
	if err != nil {
		log.Error("Error reading  record from db err:", err)
		return res, fmt.Errorf("Error reading record from db err: %s", err)
	}

	if err = listRes.All(ctx, &res); err != nil {
		log.Error(err)
		return res, fmt.Errorf("Error reading recovery record from db err: %s", err)
	}

	log.Info("The account data is: ", res)
	defer listRes.Close(ctx)

	return res, nil
}

// deleteAccount deletes an account entry
func deleteRecoveryRecord(userId string, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.D{{"recordType", Recovery}, {"userId", userId}}
	_, err := todoCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("Failed to delete recovery record ", err)
		return err
	}
	return nil
}

// writeShare write a keyShare to mongoDB from a trusted MPC dealer
func writePaillierKey(paillierKey PaillierKey, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	_, err := todoCollection.InsertOne(ctx, paillierKey)
	if err != nil {
		log.Error("failed to add BasicTx ", err)
		return err
	}
	return nil
}

// readRecoveryRecords retrieve all records of recovery
func readPaillierKeys(userId, id string, todoCollection *mongo.Collection) ([]PaillierKey, error) {
	var res []PaillierKey
	filter := bson.M{"uniqueId": id, "userId": userId}

	ctx := context.Background()
	listRes, err := todoCollection.Find(ctx, filter)
	if err == mongo.ErrNoDocuments {
		log.Info("No paillier key found:")
		return res, nil
	}

	if err != nil {
		log.Error("Error reading  paillier from db err:", err)
		return res, fmt.Errorf("Error reading record from db err: %s", err)
	}

	if err = listRes.All(ctx, &res); err != nil {
		log.Error(err)
		return res, fmt.Errorf("Error reading paillier key record from db err: %s", err)
	}

	defer listRes.Close(ctx)

	return res, nil
}

// deletePaillierKeys deletes an paillier keys entry
func deletePaillierKeys(userId, id string, todoCollection *mongo.Collection) error {
	ctx := context.Background()
	filter := bson.M{"uniqueId": id, "userId": userId}
	_, err := todoCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Error("failed to delete paillier keys ", err)
		return err
	}
	return nil
}

// readRecoveryRecords retrieve all records of recovery
func readRandomPaillierKeys(todoCollection *mongo.Collection) ([]PaillierKey, error) {
	var res []PaillierKey
	ctx := context.Background()
	pipeline := []bson.M{{"$sample": bson.M{"size": 3}}}
	listRes, err := todoCollection.Aggregate(ctx, pipeline)
	if err == mongo.ErrNoDocuments {
		log.Info("No paillier key found:")
		return res, nil
	}

	if err != nil {
		log.Error("Error reading  paillier from db err:", err)
		return res, fmt.Errorf("Error reading record from db err: %s", err)
	}

	if err = listRes.All(ctx, &res); err != nil {
		log.Error(err)
		return res, fmt.Errorf("Error reading paillier key record from db err: %s", err)
	}

	defer listRes.Close(ctx)

	return res, nil
}
