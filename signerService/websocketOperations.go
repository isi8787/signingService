package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	ep "bitbucket.org/carsonliving/cryptographymodules/ecdsaoperations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

// default value for websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//Only temporary for testing
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// wsHandler is function for managing websocket connection endpoint
// for managing ECDSA MPC signing rounds
func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()

	userId := c.Param("userId")
	blockchainId := c.Param("blockchainId")
	accountName := c.Param("accountName")

	//userId := mux.Vars(r)["userId"]
	//blockchainId := mux.Vars(r)["blockchainId"]
	//accountName := mux.Vars(r)["accountName"]
	//messageHash := mux.Vars(r)["messageHash"]
	messageHash := c.Request.URL.Query().Get("messageHash")

	log.Info("Websocket: ", userId, blockchainId, messageHash)

	//standarize hash depending on blockchain specification
	hashBytes, err := prepareHash(messageHash, blockchainId)
	if err != nil {
		log.Error("Error generating hash bytes:", err)
	}

	// initialize values for processing rounds
	var stateJson, round1JSON, round2JSON, round3JSON, round4JSON, round5JSON, round6JSON string

	//get key share collection
	var DB *mongo.Client = ConnectDB()
	keyShareCollection := DB.Database(MongoDatabase).Collection("KeyShareCollection")

	//retrieve secrete data from keyvault to begin signign rounds
	share, err := readECDSAShare(userId, blockchainId, accountName, keyShareCollection)
	if err != nil {
		log.Error("Error reading share err:", err)
	}

	//get participant id used during MPC key generation and distribution
	participantId, ok := os.LookupEnv("PARTICIPANTID")
	if !ok {
		log.Error("missing environment variable: PARTICIPANTID")
		return
	}

	broadcast1Map := make(map[string]string)
	broadcast1Map[participantId] = ""

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var response SigningRounds
		var signingMessage SigningRounds

		// decode websocket message into standard structure
		err = json.Unmarshal(message, &signingMessage)
		if err != nil {
			log.Error("Signing Message decode:", err)
		}

		//prepare standard structure for processing all MPC signing rounds
		broadcast1Map[signingMessage.Identifier] = signingMessage.Message
		pId, _ := strconv.Atoi(signingMessage.Identifier)
		signers := []int{1, pId}

		//Select message operation depending on signign rounds
		switch op := signingMessage.Round; op {
		case "round1":
			log.Info("Operation: ", op)
			round1JSON, stateJson, err = WsSignerService.WSPerformECDSARound1(ep.ECDSAParticipant(share.ShareData), signers)
			if err != nil {
				log.Error("Error round1: ", err)
			}
			broadcast1Map[participantId] = round1JSON
			response = SigningRounds{Identifier: participantId, Round: "round2", Message: round1JSON}
		case "round2":
			log.Info("Operation: ", op)
			round2JSON, stateJson, err = WsSignerService.WSPerformECDSARound2(ep.ECDSAParticipant(share.ShareData), stateJson, broadcast1Map, signers)
			if err != nil {
				log.Error("Error round 2: ", err, ", msg:", messageHash, ", userId: ", userId)
			}
			broadcast1Map[participantId] = round2JSON
			response = SigningRounds{Identifier: participantId, Round: "round3", Message: round2JSON}
		case "round3":
			log.Info("Operation: ", op)
			round3JSON, stateJson, err = WsSignerService.WSPerformECDSARound3(ep.ECDSAParticipant(share.ShareData), stateJson, broadcast1Map, signers)
			if err != nil {
				log.Error("Error round 2: ", err, ", msg:", messageHash, ", userId: ", userId)
			}
			broadcast1Map[participantId] = round3JSON
			response = SigningRounds{Identifier: participantId, Round: "round4", Message: round3JSON}
		case "round4":
			log.Info("Operation: ", op)
			round4JSON, stateJson, err = WsSignerService.WSPerformECDSARound4(ep.ECDSAParticipant(share.ShareData), stateJson, broadcast1Map, signers)
			if err != nil {
				log.Error("Error round 2: ", err, ", msg:", messageHash, ", userId: ", userId)
			}
			broadcast1Map[participantId] = round4JSON
			response = SigningRounds{Identifier: participantId, Round: "round5", Message: round4JSON}
		case "round5":
			log.Info("Operation: ", op)
			round5JSON, stateJson, err = WsSignerService.WSPerformECDSARound5(ep.ECDSAParticipant(share.ShareData), stateJson, broadcast1Map, signers)
			if err != nil {
				log.Error("Error round 2: ", err, ", msg:", messageHash, ", userId: ", userId)
			}
			broadcast1Map[participantId] = round5JSON
			response = SigningRounds{Identifier: participantId, Round: "round6", Message: round5JSON}
		case "round6":
			log.Info("Operation: ", op)
			round6JSON, stateJson, err = WsSignerService.WSPerformECDSARound6(ep.ECDSAParticipant(share.ShareData), stateJson, broadcast1Map, hashBytes, signers)
			if err != nil {
				log.Error("Error round 2: ", err, ", msg:", messageHash, ", userId: ", userId)
			}
			broadcast1Map[participantId] = round6JSON
			response = SigningRounds{Identifier: participantId, Round: "signature", Message: round6JSON}
		}

		//prepare standard response to transmit back over websocket to
		// other MPC signing participant
		responseJSON, err := json.Marshal(response)
		err = conn.WriteMessage(mt, responseJSON)
		if err != nil {
			log.Error("Error sending message:", err)
		}
	}
}

// prepareHash prepares the message has according to blockchain specification and return a byte array
// that can be used in signing process
func prepareHash(messageHash, blockchainId string) ([]byte, error) {

	switch op := blockchainId; op {
	case "ETH", "BNB", "MATIC":
		hash := common.HexToHash(messageHash)
		return hash[:], nil

	case "AVAX":
		hash := common.HexToHash(messageHash)
		return hash[:], nil

	case "BTC":
		hash, err := hex.DecodeString(messageHash)
		if err != nil {
			return nil, err
		}
		return hash[:], nil

	}

	return nil, nil
}
